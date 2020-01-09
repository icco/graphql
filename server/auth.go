package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/icco/graphql"
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
	Scope     string   `json:"scope"`
	Audience  []string `json:"aud,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	Id        string   `json:"jti,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
}

func (c CustomClaims) toStandard() jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  c.Audience[0],
		ExpiresAt: c.ExpiresAt,
		Id:        c.Id,
		IssuedAt:  c.IssuedAt,
		Issuer:    c.Issuer,
		NotBefore: c.NotBefore,
		Subject:   c.Subject,
	}
}

func (c CustomClaims) Valid() error {
	return c.toStandard().Valid()
}

// APIKeyMiddleware is an auth middleware. If user is coming in via api key
// header, use that as your auth.
func APIKeyMiddleware(next http.Handler) http.Handler {
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

func jsonError(msg string) error {
	data, err := json.Marshal(map[string]string{"error": msg})
	if err != nil {
		log.WithError(err).Error("could not marshal json")
	}
	return fmt.Errorf("%s", data)
}

// AuthMiddleware parses the incomming authentication header and turns it into
// an attached user.
func AuthMiddleware(next http.Handler) http.Handler {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		CredentialsOptional: true,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := "https://natwelch.com"
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				log.Errorf("invalid audence: %q", token)
				return token, jsonError("Invalid audience.")
			}
			// Verify 'iss' claim
			iss := "https://icco.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				log.Errorf("invalid issuer: %q", token)
				return token, jsonError("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				msg := "cloudn't parse pem cert"
				log.WithError(err).Error(msg)
				return token, jsonError(msg)
			}

			data, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			if err != nil {
				log.WithError(err).WithField("cert", cert).Errorf("error parsing cert")
				return token, jsonError(err.Error())
			}
			return data, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	return jwtMiddleware.Handler(getUserFromToken(next))
}

func getUserFromToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := jwtmiddleware.FromAuthHeader(r)
		if err != nil {
			log.WithError(err).Error("could not get auth header")
			next.ServeHTTP(w, r)
			return
		}

		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			cert, err := getPemCert(token)
			if err != nil {
				return nil, err
			}
			return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
		})

		if err != nil {
			log.WithError(err).Error("could not get claims")
			next.ServeHTTP(w, r)
			return
		}

		log.WithField("token", token).WithField("claims", claims).Debug("the token")

		next.ServeHTTP(w, r)
	})
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://icco.auth0.com/.well-known/jwks.json")

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
		err := fmt.Errorf("Unable to find appropriate key.")
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
