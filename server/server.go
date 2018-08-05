package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/99designs/gqlgen/handler"
	"github.com/icco/graphql"
	"gopkg.in/unrolled/render.v1"
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

	// Basic router
	http.HandleFunc("/healthz", healthCheckHandler)
	http.Handle("/", handler.Playground("graphql", "/query"))
	http.Handle("/query", handler.GraphQL(
		graphql.NewExecutableSchema(graphql.New()),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			log.Print(err)
			debug.PrintStack()
			return errors.New("Panic message seen when processing request")
		}),
	))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	Renderer.JSON(w, http.StatusOK, map[string]string{
		"healthy":  "true",
		"revision": os.Getenv("GIT_REVISION"),
		"tag":      os.Getenv("GIT_TAG"),
		"branch":   os.Getenv("GIT_BRANCH"),
	})
}
