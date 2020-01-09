package main

import (
	"net/http"
	"os"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/icco/graphql"
)

var (
	// AUTH0 holds our Auth0 config constants.
	AUTH0 = map[string]string{
		"API-SECRET":   os.Getenv("AUTH0_API_SECRET"),
		"API-AUDIENCE": os.Getenv("AUTH0_API_AUDIENCE"),
		"DOMAIN":       os.Getenv("AUTH0_DOMAIN"),
	}
)

// Jwks is from https://auth0.com/docs/quickstart/backend/golang
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

// JSONWebKeys is from https://auth0.com/docs/quickstart/backend/golang
type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// CustomClaims is from https://auth0.com/docs/quickstart/backend/golang
type CustomClaims struct {
	Scope string `json:"scope"`
	jwt.StandardClaims
}

// APIKeyMiddleware is an auth middleware. If user is coming in via api key
// header, use that as your auth.
func APIKeyMiddleware(next http.handler) http.handler {
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
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware parses the incomming authentication header and turns it into
// an attached user.
func AuthMiddleware(next http.handler) http.handler {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := "https://natwelch.com"
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("Invalid audience.")
			}
			// Verify 'iss' claim
			iss := "https://natwelch.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				log.WithError(err).Error("cloudn't parse pem cert")
			}

			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	return jwtMiddleware.Handler(next)
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://natwelch.com/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

func checkScope(scope string, tokenString string) bool {
	token, _ := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		cert, err := getPemCert(token)
		if err != nil {
			return nil, err
		}
		result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		return result, nil
	})

	claims, ok := token.Claims.(*CustomClaims)

	hasScope := false
	if ok && token.Valid {
		result := strings.Split(claims.Scope, " ")
		for i := range result {
			if result[i] == scope {
				hasScope = true
			}
		}
	}

	return hasScope
}
