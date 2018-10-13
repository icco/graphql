package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/icco/graphql"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"gopkg.in/unrolled/render.v1"
	"gopkg.in/unrolled/secure.v1"
)

var (
	// Renderer is a renderer for all occasions. These are our preferred default options.
	// See:
	//  - https://github.com/unrolled/render/blob/v1/README.md
	//  - https://godoc.org/gopkg.in/unrolled/render.v1
	Renderer = render.New(render.Options{
		Charset:                   "UTF-8",
		Directory:                 "./server/views",
		DisableHTTPErrorRendering: false,
		Extensions:                []string{".tmpl", ".html"},
		IndentJSON:                false,
		IndentXML:                 true,
		Layout:                    "layout",
		RequirePartials:           true,
		Funcs:                     []template.FuncMap{{}},
	})

	dbURL = os.Getenv("DATABASE_URL")
)

func main() {
	if dbURL == "" {
		log.Panicf("DATABASE_URL is empty!")
	}

	graphql.InitDB(dbURL)
	OAuthConfig = configureOAuthClient(
		os.Getenv("OAUTH2_CLIENTID"),
		os.Getenv("OAUTH2_SECRET"),
		os.Getenv("OAUTH2_REDIRECT"))

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on http://localhost:%s", port)

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "graphql",
	})
	if err != nil {
		log.Fatalf("Failed to create the Prometheus exporter: %v", err)
	}
	view.RegisterExporter(pe)

	if os.Getenv("ENABLE_STACKDRIVER") != "" {
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:    "icco-cloud",
			MetricPrefix: "graphql",
		})
		if err != nil {
			log.Fatalf("Failed to create the Stackdriver exporter: %v", err)
		}
		defer sd.Flush()
		view.RegisterExporter(sd)
		trace.RegisterExporter(sd)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	isDev := os.Getenv("NAT_ENV") != "production"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(ContextMiddleware)

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
		r.Handle("/metrics", pe)
	})

	// Everything that does SSL only
	r.Group(func(r chi.Router) {
		r.Use(secure.New(secure.Options{
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
		}).Handler)

		r.Mount("/admin", adminRouter())

		r.Handle("/", handler.Playground("graphql", "/graphql"))
		r.Handle("/graphql", handler.GraphQL(
			graphql.NewExecutableSchema(graphql.New()),
			handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
				log.Print(err)
				debug.PrintStack()
				return errors.New("Panic message seen when processing request")
			}),
		))

		// Auth stuff
		r.HandleFunc("/login", loginHandler)
		r.HandleFunc("/logout", logoutHandler)
		r.HandleFunc("/callback", callbackHandler)
	})

	h := &ochttp.Handler{
		Handler:          r,
		IsPublicEndpoint: true,
	}
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatal("Failed to register ochttp.DefaultServerViews")
	}

	log.Fatal(http.ListenAndServe(":"+port, h))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.JSON(w, http.StatusOK, map[string]string{
		"healthy":  "true",
		"revision": os.Getenv("GIT_REVISION"),
		"tag":      os.Getenv("GIT_TAG"),
		"branch":   os.Getenv("GIT_BRANCH"),
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.HTML(w, http.StatusNotFound, "404", struct{ Title string }{Title: "404: This page could not be found"})
}
