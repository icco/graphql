package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/icco/graphql"
)

const (
	timeFormat = "2006-01-02T15:04"
)

type adminPageData struct {
	Title    string
	Posts    []*graphql.Post
	Post     *graphql.Post
	Datetime string
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Renderer.HTML(w, http.StatusOK, "admin", &adminPageData{Title: "Admin"})
	})

	r.Get("/post/new", func(w http.ResponseWriter, r *http.Request) {
		Renderer.HTML(w, http.StatusOK, "new_post", &adminPageData{
			Title:    "New Post",
			Datetime: time.Now().Format(timeFormat),
		})
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
		if len(r.Form["datetime"]) == 1 {
			// TODO: Parse datetime
			log.Printf("dt: %+v", r.Form["datetime"])
		} else {
			// TODO: raise an error
		}

		post := graphql.GeneratePost(title, text, datetime, []string{})
		_, err := graphql.CreatePost(r.Context(), *post)

		if err != nil {
			log.Printf("err: %+v", err)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})

	return r
}
