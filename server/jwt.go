package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/icco/graphql"
)

// AdminOnly is a middleware that makes sure the logged in user is an admin, or
// 403.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authBackend := InitJWTAuthenticationBackend()
		var user *graphql.User

		// If error, we couldn't parse session.
		allowed := false
		token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			} else {
				return authBackend.PublicKey, nil
			}
		})

		if err != nil {
			log.WithError(err).Error("JWT parsing error")
		} else {
			if token.Valid {
				user, err = graphql.GetUser(r.Context(), token.UserID)
				if err != nil {
					appErrorf(w, err, "could not upsert user: %v", err)
					return
				}

				allowed = user.Role == "admin"
			}
		}

		if !allowed {
			log.Printf("User could not login: %+v", user)
			http.Error(w, http.StatusText(403), 403)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ContextMiddleware gets the current user in the session and stores in the
// current context.
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := SessionStore.Get(r, defaultSessionID)

		apikey := ""
		if len(r.Header["Authorization"]) > 0 {
			apikey = r.Header["Authorization"][0]
		}

		// Allow unauthenticated users in
		if err != nil || session == nil || (session.Values[googleProfileSessionKey] == nil && apikey == "") {
			next.ServeHTTP(w, r)
			return
		}

		if apikey != "" {
			user, err := graphql.GetUserByAPIKey(r.Context(), apikey)
			if err != nil {
				appErrorf(w, err, "could not get user by apikey: %v", err)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), graphql.UserCtxKey, user)
			r = r.WithContext(ctx)
		} else {
			// get the user from the database
			profile := session.Values[googleProfileSessionKey].(*graphql.User)
			if profile.ID != "" {
				user, err := graphql.GetUser(r.Context(), profile.ID)
				if err != nil {
					appErrorf(w, err, "could not upsert user: %v", err)
					return
				}

				// put it in context
				ctx := context.WithValue(r.Context(), graphql.UserCtxKey, user)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}
