package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/icco/writing-be/models"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
)

var schema *graphql.Schema

func addDefaultHeaders(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

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

	server := http.NewServeMux()
	server.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}))
	server.Handle("/graphql", &relay.Handler{Schema: schema})
	server.HandleFunc("/healthz", healthCheckHandler)

	headedRouter := addDefaultHeaders(server)
	loggedRouter := handlers.LoggingHandler(os.Stdout, headedRouter)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggedRouter))
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
