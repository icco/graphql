package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/icco/writing"
)

func main() {
	http.Handle("/", handler.Playground("writing", "/query"))
	http.Handle("/query", handler.GraphQL(
		writing.NewExecutableSchema(writing.New()),
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
