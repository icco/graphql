package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/icco/writing-be/models"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"gopkg.in/unrolled/secure.v1"
)

var schema *graphql.Schema

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}
	log.Printf("Starting up on %s", port)

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Panicf("DATABASE_URL is empty!")
	}
	log.Printf("Got DB URL %s", dbUrl)

	models.InitDB(dbUrl)
	log.Printf("Got passed db line")

	schema = graphql.MustParseSchema(models.Schema, &models.Resolver{})

	server := chi.NewRouter()

	// Common suggested middleware
	server.Use(middleware.RequestID)
	server.Use(middleware.RealIP)
	server.Use(middleware.Logger)
	server.Use(middleware.Recoverer)

	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
	})

	server.Use(secureMiddleware.Handler)

	server.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))
	server.Get("/healthz", healthCheckHandler)

	relayHandler := &relay.Handler{Schema: schema}
	server.Get("/graphql", relayHandler.ServeHTTP)
	server.Post("/graphql", relayHandler.ServeHTTP)
	server.Options("/graphql", relayHandler.ServeHTTP)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

type healthRespJSON struct {
	Healthy string `json:"healthy"`
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := healthRespJSON{
		Healthy: "true",
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

var page = []byte(`
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.0/graphiql.css" />
		<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.0/graphiql.js"></script>
	</head>
	<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
		<div id="graphiql" style="height: 100vh;">Loading...</div>
		<script>
			function graphQLFetcher(graphQLParams) {
				return fetch("/graphql", {
					method: "post",
					body: JSON.stringify(graphQLParams),
					credentials: "include",
				}).then(function (response) {
					return response.text();
				}).then(function (responseBody) {
					try {
						return JSON.parse(responseBody);
					} catch (error) {
						return responseBody;
					}
				});
			}

			ReactDOM.render(
				React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
				document.getElementById("graphiql")
			);
		</script>
	</body>
</html>
`)
