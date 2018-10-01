package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/icco/graphql"
)

type adminPageData struct {
	Title string
	Posts []*graphql.Post
	Post  *graphql.Post
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Renderer.HTML(w, http.StatusOK, "admin", &adminPageData{Title: "Admin"})
	})
	r.Get("/post/new", func(w http.ResponseWriter, r *http.Request) {
		Renderer.HTML(w, http.StatusOK, "new_post", &adminPageData{Title: "New Post"})
	})
	r.Post("/post/new", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		title := ""
		if len(r.Form["title"]) == 1 {
			title = r.Form["title"][0]
		} else {
			// TODO: raise an error
		}

		text := ""
		if len(r.Form["text"]) == 1 {
			text = r.Form["text"][0]
		} else {
			// TODO: raise an error
		}

		datetime := time.Now()

		post := graphql.GeneratePost(title, text, datetime, []string{})
		_, err := graphql.CreatePost(r.Context(), *post)

		if err != nil {
			log.Printf("err: %+v", err)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})

	return r
}
