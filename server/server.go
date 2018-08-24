package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/icco/graphql"
	"github.com/rs/cors"
	"gopkg.in/unrolled/render.v1"
	"gopkg.in/unrolled/secure.v1"
)

// Renderer is a renderer for all occasions. These are our preferred default options.
// See:
//  - https://github.com/unrolled/render/blob/v1/README.md
//  - https://godoc.org/gopkg.in/unrolled/render.v1
var Renderer = render.New(render.Options{
	Charset:                   "UTF-8",
	Directory:                 "views",
	DisableHTTPErrorRendering: false,
	Extensions:                []string{".tmpl", ".html"},
	IndentJSON:                false,
	IndentXML:                 true,
	Layout:                    "layout",
	RequirePartials:           true,
})

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Panicf("DATABASE_URL is empty!")
	}
	log.Printf("Got DB URL %s", dbUrl)

	graphql.InitDB(dbUrl)

	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on %s", port)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.New(cors.Options{
		AllowCredentials:   true,
		OptionsPassthrough: true,
	}).Handler)

	r.Use(secure.New(secure.Options{
		HostsProxyHeaders:  []string{"X-Forwarded-Host"},
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		IsDevelopment:      false,
	}).Handler)

	r.Get("/healthz", healthCheckHandler)

	r.Handle("/", handler.Playground("graphql", "/graphql"))
	r.Handle("/graphql", handler.GraphQL(
		graphql.NewExecutableSchema(graphql.New()),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			log.Print(err)
			debug.PrintStack()
			return errors.New("Panic message seen when processing request")
		}),
	))

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.JSON(w, http.StatusOK, map[string]string{
		"healthy":  "true",
		"revision": os.Getenv("GIT_REVISION"),
		"tag":      os.Getenv("GIT_TAG"),
		"branch":   os.Getenv("GIT_BRANCH"),
	})
}
