package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
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
	ID        string   `json:"jti,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	NotBefore int64    `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
}

func (c CustomClaims) toStandard() jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  c.Audience,
		ExpiresAt: c.ExpiresAt,
		Id:        c.ID,
		IssuedAt:  c.IssuedAt,
		Issuer:    c.Issuer,
		NotBefore: c.NotBefore,
		Subject:   c.Subject,
	}
}

// Valid is required for conformance to jwt.Claims.
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
				log.Errorf("invalid audence: %q", token.Raw)
				return token, jsonError("Invalid audience.")
			}
			// Verify 'iss' claim
			iss := "https://icco.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				log.Errorf("invalid issuer: %q", token.Raw)
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
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
			log.Errorf("error with auth: %q", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, err)
			return
		},
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
			log.WithError(err).Debug("could not get claims")
			next.ServeHTTP(w, r)
			return
		}

		log.WithField("token", token).WithField("claims", claims).Debug("the token")

		if claims.Subject != "" {
			user, err := graphql.GetUser(r.Context(), claims.Subject)
			if err != nil {
				log.WithError(err).WithField("claims", claims).Error("could not get user")
			} else {
				// put it in context
				ctx := graphql.WithUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
		}

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

	for _, k := range jwks.Keys {
		if token.Header["kid"] == k.Kid {
			cert = fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", k.X5c[0])
		}
	}

	if cert == "" {
		err := fmt.Errorf("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}
