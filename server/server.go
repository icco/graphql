package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/monitoredresource"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/99designs/gqlgen-contrib/gqlapollotracing"
	"github.com/99designs/gqlgen-contrib/gqlopencensus"
	gql "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/icco/graphql"
	sdLogging "github.com/icco/logrus-stackdriver-formatter"
	"github.com/unrolled/render"
	"github.com/unrolled/secure"
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

	var trcr gql.Tracer
	trcr = gqlapollotracing.NewTracer()

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
			log.Fatalf("Failed to create the Stackdriver exporter: %v", err)
		}
		defer sd.Flush()

		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{
			DefaultSampler: trace.ProbabilitySampler(0.1),
		})

		trcr = gqlopencensus.New()
	}

	isDev := os.Getenv("NAT_ENV") != "production"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultCompress)
	r.Use(sdLogging.LoggingMiddleware(log))

	r.Use(cors.New(cors.Options{
		AllowCredentials:   true,
		OptionsPassthrough: true,
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:     []string{"Link"},
		MaxAge:             300, // Maximum value not ignored by any of major browsers
	}).Handler)

	r.NotFound(notFoundHandler)

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
		r.Options("/photo/new", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(""))
		})
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
		r.Use(AuthMiddleware)

		r.Get("/cron", cronHandler)
		r.Handle("/", handler.Playground("graphql", "/graphql"))
		r.Handle("/graphql", handler.GraphQL(
			graphql.NewExecutableSchema(graphql.New()),
			handler.RecoverFunc(func(ctx context.Context, intErr interface{}) error {
				err, ok := intErr.(error)
				if ok {
					log.WithError(err).Error("Error seen during graphql")
				}
				return errors.New("fatal error seen while processing request")
			}),
			handler.CacheSize(512),
			handler.RequestMiddleware(GqlLoggingMiddleware),
			handler.RequestMiddleware(gqlapollotracing.RequestMiddleware()),
			handler.Tracer(trcr),
		))

		r.Post("/photo/new", photoUploadHandler)
	})

	h := &ochttp.Handler{
		Handler:     r,
		Propagation: &propagation.HTTPFormat{},
	}
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatal("Failed to register ochttp.DefaultServerViews")
	}

	log.Fatal(http.ListenAndServe(":"+port, h))
}

// GqlLoggingMiddleware is a middleware for gqlgen that logs all gql requests to debug.
func GqlLoggingMiddleware(ctx context.Context, next func(ctx context.Context) []byte) []byte {
	rctx := gql.GetRequestContext(ctx)

	// We do this because RequestContext has fields that can't be easily
	// serialized in json, and we don't care about them.
	subsetContext := map[string]interface{}{
		"query":      rctx.RawQuery,
		"variables":  rctx.Variables,
		"extensions": rctx.Extensions,
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
						log.WithError(err).Printf("Error saving post")
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
		"error": "500: An internal server error occured",
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
