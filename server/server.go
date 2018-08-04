package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/99designs/gqlgen/handler"
	"github.com/icco/writing"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Panicf("DATABASE_URL is empty!")
	}
	log.Printf("Got DB URL %s", dbUrl)

	writing.InitDB(dbUrl)

	// Basic router
	http.Handle("/", handler.Playground("writing", "/query"))
	http.Handle("/query", handler.GraphQL(
		writing.NewExecutableSchema(writing.New()),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) error {
			log.Print(err)
			debug.PrintStack()
			return errors.New("Panic message seen when processing request")
		}),
	))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
