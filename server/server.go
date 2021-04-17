package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	gql "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/icco/graphql"
	"github.com/icco/gutil/logging"
	"github.com/unrolled/render"
	"github.com/unrolled/secure"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

var (
	// Renderer is a renderer for all occasions. These are our preferred default options.
	// See:
	//  - https://github.com/unrolled/render/blob/v1/README.md
	//  - https://godoc.org/gopkg.in/unrolled/render.v1
	Renderer = render.New(render.Options{
		Charset:                   "UTF-8",
		Directory:                 "./server/views",
		DisableHTTPErrorRendering: true,
		Extensions:                []string{".tmpl", ".html"},
		Funcs:                     []template.FuncMap{{}},
		IndentJSON:                false,
		IndentXML:                 true,
		Layout:                    "layout",
		RequirePartials:           true,
	})

	dbURL = os.Getenv("DATABASE_URL")
	log   = logging.Must(logging.NewLogger(graphql.AppName))

	// GCPProjectID is the project ID where we should send errors.
	GCPProjectID = "icco-cloud"
)

func main() {
	if dbURL == "" {
		log.Fatal("DATABASE_URL is empty!")
	}

	if _, err := graphql.InitDB(dbURL); err != nil {
		log.Fatalw("Init DB", zap.Error(err))
	}

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Infow("Starting up", "host", fmt.Sprintf("http://localhost:%s", port))

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		labels := &stackdriver.Labels{}
		labels.Set("app", graphql.AppName, "The name of the current app.")
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               GCPProjectID,
			MonitoredResource:       monitoredresource.Autodetect(),
			DefaultMonitoringLabels: labels,
			OnError: func(err error) {
				log.Errorw("couldn't upload to stackdriver", zap.Error(err))
			},
		})

		if err != nil {
			log.Fatalw("failed to create the Stackdriver exporter", zap.Error(err))
		}
		defer sd.Flush()

		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.ProbabilitySampler(0.1),
		})

	}

	isDev := os.Getenv("NAT_ENV") != "production"

	cache, err := graphql.NewCache()
	if err != nil {
		log.Fatalw("could not connect to cache", zap.Error(err))
	}

	gh := handler.New(graphql.NewExecutableSchema(graphql.New()))
	gh.AddTransport(transport.Websocket{KeepAlivePingInterval: 10 * time.Second})
	gh.AddTransport(transport.Options{})
	gh.AddTransport(transport.GET{})
	gh.AddTransport(transport.POST{})
	gh.AddTransport(transport.MultipartForm{})

	gh.SetQueryCache(lru.New(1000))

	gh.Use(extension.AutomaticPersistedQuery{Cache: cache})
	gh.Use(extension.Introspection{})

	gh.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := gql.DefaultErrorPresenter(ctx, e)

		log.Warnw("graphql request error", zap.Error(e))
		if strings.Contains(e.Error(), "forbidden") {
			return gqlerror.Errorf("forbidden: not a valid user")
		}

		return err
	})
	gh.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		if e, ok := err.(error); !ok {
			log.Errorw("graphql fatal request error", "error", err)
		} else {
			log.Errorw("graphql fatal request error", zap.Error(e))
		}

		return fmt.Errorf("Internal server error!")
	})

	gh.AroundResponses(GqlLoggingMiddleware)
	// TODO: Add this back once gqlgen-contrib supports it.
	//handler.Tracer(gqlopencensus.New())

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(logging.Middleware(log.Desugar(), GCPProjectID))

	crs := cors.New(cors.Options{
		AllowCredentials:   true,
		OptionsPassthrough: false,
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "x-apollo-tracing"},
		ExposedHeaders:     []string{"Link"},
		MaxAge:             300, // Maximum value not ignored by any of major browsers
	})
	r.NotFound(notFoundHandler)
	r.Use(crs.Handler)

	// Stuff that does not ssl redirect
	r.Group(func(r chi.Router) {
		r.Use(secure.New(secure.Options{
			BrowserXssFilter:   true,
			ContentTypeNosniff: true,
			FrameDeny:          true,
			HostsProxyHeaders:  []string{"X-Forwarded-Host"},
			IsDevelopment:      isDev,
			SSLProxyHeaders:    map[string]string{"X-Forwarded-Proto": "https"},
		}).Handler)

		r.Get("/healthz", healthCheckHandler)
	})

	// Everything that does SSL only
	sslOnly := secure.New(secure.Options{
		BrowserXssFilter:     true,
		ContentTypeNosniff:   true,
		FrameDeny:            true,
		HostsProxyHeaders:    []string{"X-Forwarded-Host"},
		IsDevelopment:        isDev,
		SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
		SSLRedirect:          !isDev,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		STSSeconds:           315360000,
	}).Handler

	r.Group(func(r chi.Router) {
		r.Use(sslOnly)
		r.Use(APIKeyMiddleware)
		r.Use(AuthMiddleware)

		r.Get("/cron", cronHandler)
		r.Handle("/", playground.Handler("graphql", "/graphql"))
		r.Handle("/graphql", gh)

		r.Post("/photo/new", photoUploadHandler)
	})

	h := &ochttp.Handler{
		Handler:     r,
		Propagation: &propagation.HTTPFormat{},
	}
	if err := view.Register([]*view.View{
		ochttp.ServerRequestCountView,
		ochttp.ServerResponseCountByStatusCode,
	}...); err != nil {
		log.Fatal("Failed to register ochttp.DefaultServerViews")
	}

	log.Fatal(http.ListenAndServe(":"+port, h))
}

// GqlLoggingMiddleware is a middleware for gqlgen that logs all gql requests to debug.
func GqlLoggingMiddleware(ctx context.Context, next gql.ResponseHandler) *gql.Response {
	rctx := gql.GetOperationContext(ctx)

	// We do this because RequestContext has fields that can't be easily
	// serialized in json, and we don't care about them.
	subsetContext := map[string]interface{}{
		"query":     rctx.RawQuery,
		"variables": rctx.Variables,
		"name":      rctx.OperationName,
		"stats":     rctx.Stats,
	}

	log.Debugw("request gql", "gql", subsetContext)

	return next(ctx)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.JSON(w, http.StatusOK, map[string]string{
		"healthy": "true",
	})
}

func cronHandler(w http.ResponseWriter, r *http.Request) {
	go func(ctx context.Context) {
		var posts []*graphql.Post
		var err error
		perPage := 10

		for i := 0; err == nil || len(posts) > 0; i += perPage {
			posts, err = graphql.Posts(ctx, perPage, i)
			if err == nil {
				for _, p := range posts {
					err = p.Save(ctx)
					if err != nil {
						log.Errorw("error saving post", zap.Error(err))
					}
				}
			}
		}
	}(context.Background())

	err := Renderer.JSON(w, http.StatusOK, map[string]string{
		"cron": "ok",
	})
	if err != nil {
		log.Errorw("could not render json", zap.Error(err))
	}
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {

	err := Renderer.JSON(w, http.StatusInternalServerError, map[string]string{
		"error": "500: An internal server error occurred",
	})
	if err != nil {
		log.Errorw("could not render json", zap.Error(err))
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := Renderer.JSON(w, http.StatusNotFound, map[string]string{
		"error": "404: This page could not be found",
	})
	if err != nil {
		log.Errorw("could not render json", zap.Error(err))
	}
}
