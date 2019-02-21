package main

import (
	"net/http"
	"os"

	"github.com/auth0-community/go-auth0"
	"github.com/icco/graphql"
	jose "gopkg.in/square/go-jose.v2"
)

var (
	// AUTH0 holds our Auth0 config constants.
	AUTH0 = map[string]string{
		"API-SECRET":   os.Getenv("AUTH0_API_SECRET"),
		"API-AUDIENCE": os.Getenv("AUTH0_API_AUDIENCE"),
		"DOMAIN":       os.Getenv("AUTH0_DOMAIN"),
	}
)

// AuthMiddleware parses the incomming authentication header and turns it into
// an attached user.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// API Key dropout
		if r.Header.Get("X-API-AUTH") != "" {
			apikey := r.Header.Get("X-API-AUTH")
			user, err := graphql.GetUserByAPIKey(r.Context(), apikey)
			if err != nil {
				log.WithError(err).Error("could not get user by apikey")
				http.Error(w, `{"error": "could not get a user with that API key"}`, http.StatusBadRequest)
				return
			}

			// put it in context
			ctx := graphql.WithUser(r.Context(), user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		secretProvider := auth0.NewJWKClient(auth0.JWKClientOptions{URI: AUTH0["DOMAIN"] + "/.well-known/jwks.json"}, nil)
		audience := []string{AUTH0["API-AUDIENCE"]}
		configuration := auth0.NewConfiguration(secretProvider, audience, AUTH0["DOMAIN"]+"/", jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		token, err := validator.ValidateRequest(r)

		// This also returns an error if no token
		if err != nil {
			if err.Error() == "Token not found" {
				next.ServeHTTP(w, r)
				return
			}

			log.WithField("auth", AUTH0).WithError(err).Error("token is not valid")
			http.Error(w, `{"error": "Error reading auth token"}`, http.StatusBadRequest)
			return
		}

		claims := map[string]interface{}{}
		err = validator.Claims(r, token, &claims)
		if err != nil {
			log.WithError(err).Error("Claims are not valid")
			http.Error(w, http.StatusText(403), 403)
			return
		}

		// According to Auth0, the sub claim is what we're supposed to use to identify a unique user... so here we go!
		userid, ok := claims["sub"].(string)
		if ok {
			user, err := graphql.GetUser(r.Context(), userid)
			if err != nil {
				log.WithError(err).WithField("claims", claims).Error("could not get user")
			} else {
				// put it in context
				ctx := graphql.WithUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
		return
	})
}
