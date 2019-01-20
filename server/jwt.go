package main

import (
	"net/http"
	"os"

	"github.com/auth0-community/auth0"
	jose "gopkg.in/square/go-jose.v2"
)

var (
	AUTH0 = map[string]string{
		"API-SECRET":   os.Getenv("AUTH0_API_SECRET"),
		"API-AUDIENCE": os.Getenv("AUTH0_API_AUDIENCE"),
		"DOMAIN":       os.Getenv("AUTH0_DOMAIN"),
	}
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secretProvider := auth0.NewKeyProvider([]byte(AUTH0["API-SECRET"]))
		audience := []string{AUTH0["API-AUDIENCE"]}

		configuration := auth0.NewConfiguration(secretProvider, audience, AUTH0["DOMAIN"], jose.HS256)
		validator := auth0.NewValidator(configuration)

		token, err := validator.ValidateRequest(r)

		if err != nil {
			log.WithError(err).Error("Token is not valid")
			http.Error(w, http.StatusText(403), 403)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
