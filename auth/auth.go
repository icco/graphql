package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// Middleware decodes the share session cookie and packs the session into
// context. It is based off of https://gqlgen.com/recipes/authentication/
func Middleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("auth-cookie")

			// Allow unauthenticated users in
			if err != nil || c == nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := validateAndGetUser(c)
			if err != nil {
				http.Error(w, "Invalid cookie", http.StatusForbidden)
				return
			}

			// put it in context
			ctx := context.WithValue(r.Context(), "user", user)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// CallbackHandler is a handler that will be proxied by the frontend to the
// backend to cache the auth stuff.
func CallbackHandler(w http.ResponseWriter, r *http.Request) {

	domain := os.Getenv("AUTH0_DOMAIN")

	conf := &oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("http://localhost:3000/callback"),
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/authorize", domain),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", domain),
		},
	}
	state := r.URL.Query().Get("state")

	stateParam := r.Context().Value("state")

	if state != stateParam {
		log.Printf("Invalid state param: %+v != %+v", state, stateParam)
		http.Error(w, "Invalid state parameter", http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")

	token, err := conf.Exchange(context.TODO(), code)
	if err != nil {
		log.Printf("getting token: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Getting now the userInfo
	client := conf.Client(context.TODO(), token)
	resp, err := client.Get("https://" + domain + "/userinfo")
	if err != nil {
		log.Printf("getting user: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var profile map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	context.WithValue(ctx, "id_token", token.Extra("id_token"))
	context.WithValue(ctx, "access_token", token.AccessToken)
	context.WithValue(ctx, "profile", profile)

	// Redirect to logged in page
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
