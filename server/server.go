package main

import (
	"context"
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
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/icco/graphql"
	sdLogging "github.com/icco/logrus-stackdriver-formatter"
	"github.com/unrolled/render"
	"github.com/unrolled/secure"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
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

	log = graphql.InitLogging()
)

func main() {
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is empty!")
	}

	_, err := graphql.InitDB(dbURL)
	if err != nil {
		log.Fatalf("Init DB: %+v", err)
	}

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on http://localhost:%s", port)

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		labels := &stackdriver.Labels{}
		labels.Set("app", "graphql", "The name of the current app.")
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               "icco-cloud",
			MonitoredResource:       monitoredresource.Autodetect(),
			DefaultMonitoringLabels: labels,
			OnError: func(err error) {
				log.WithError(err).Error("couldn't upload to stackdriver")
			},
		})

		if err != nil {
			log.WithError(err).Fatal("failed to create the Stackdriver exporter")
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
		log.WithError(err).Fatal("could not connect to cache")
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

	gh.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		log.WithError(err).Error("graphql request error")
		if strings.Contains(err.Error(), "forbidden") {
			return gqlerror.Errorf("forbidden: not a valid user")
		}

		return gqlerror.Errorf("error seen while processing request")
	})

	gh.AroundResponses(GqlLoggingMiddleware)
	// TODO: Add this back once gqlgen-contrib supports it.
	//handler.Tracer(gqlopencensus.New())

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(sdLogging.LoggingMiddleware(log))

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

	log.WithField("gql", subsetContext).Debug("request gql")

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
						log.WithError(err).Errorf("error saving post")
					}
				}
			}
		}
	}(context.Background())

	err := Renderer.JSON(w, http.StatusOK, map[string]string{
		"cron": "ok",
	})
	if err != nil {
		log.WithError(err).Error("could not render json")
	}
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {

	err := Renderer.JSON(w, http.StatusInternalServerError, map[string]string{
		"error": "500: An internal server error occurred",
	})
	if err != nil {
		log.WithError(err).Error("could not render json")
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	err := Renderer.JSON(w, http.StatusNotFound, map[string]string{
		"error": "404: This page could not be found",
	})
	if err != nil {
		log.WithError(err).Error("could not render json")
	}
}
