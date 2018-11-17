package main

import (
	"fmt"
	"net/http"
	"strconv"
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
	Drafts   []*graphql.Post
	Post     *graphql.Post
	Datetime string
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := graphql.AllPosts(r.Context(), false)
		if err != nil {
			log.Printf("Error getting posts: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		drafts, err := graphql.AllPosts(r.Context(), true)
		if err != nil {
			log.Printf("Error getting drafts: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		Renderer.HTML(w, http.StatusOK, "admin", &adminPageData{
			Title:  "Admin",
			Posts:  posts,
			Drafts: drafts,
		})
	})

	r.Get("/post/new", func(w http.ResponseWriter, r *http.Request) {
		Renderer.HTML(w, http.StatusOK, "new_post", &adminPageData{
			Title:    "New Post",
			Datetime: time.Now().Format(timeFormat),
		})
	})

	r.Post("/post/new", func(w http.ResponseWriter, r *http.Request) {
		var err error
		r.ParseForm()

		title := ""
		if len(r.Form["title"]) == 1 {
			title = r.Form["title"][0]
		}

		text := ""
		if len(r.Form["text"]) == 1 {
			text = r.Form["text"][0]
		} else {
			// TODO: raise an error
		}

		datetime := time.Now()
		if len(r.Form["datetime"]) == 1 {
			datetime, err = time.Parse(timeFormat, r.Form["datetime"][0])
			if err != nil {
				log.Printf("Error parsing time: %+v", err)
				http.Error(w, "Error parsing time.", http.StatusInternalServerError)
				return
			}
		}

		draft := len(r.Form["draft"]) == 1 && r.Form["draft"][0] == "on"

		post := graphql.GeneratePost(r.Context(), title, text, datetime, []string{}, draft)
		err = post.Save(r.Context())

		if err != nil {
			log.Printf("err: %+v", err)
		}

		http.Redirect(w, r, "/admin/", http.StatusFound)
	})

	r.Get("/edit/{post_id}", func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(chi.URLParam(r, "post_id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing int: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		post, err := graphql.GetPost(r.Context(), postID)
		if err != nil {
			log.Printf("Error editing post: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		Renderer.HTML(w, http.StatusOK, "edit_post", &adminPageData{
			Title:    fmt.Sprintf("Edit Post #%v", post.ID),
			Post:     post,
			Datetime: post.Datetime.Format(timeFormat),
		})
	})

	r.Post("/edit/{post_id}", func(w http.ResponseWriter, r *http.Request) {
		var err error
		r.ParseForm()

		postID, err := strconv.ParseInt(chi.URLParam(r, "post_id"), 10, 64)
		if err != nil {
			log.Printf("Error parsing int: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		post, err := graphql.GetPost(r.Context(), postID)
		if err != nil {
			log.Printf("Error editing post: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(r.Form["title"]) == 1 {
			post.Title = r.Form["title"][0]
		}

		if len(r.Form["text"]) == 1 {
			post.Content = r.Form["text"][0]
		}

		if len(r.Form["datetime"]) == 1 {
			datetime, err := time.Parse(timeFormat, r.Form["datetime"][0])
			if err != nil {
				log.Printf("Error parsing time: %+v", err)
				http.Error(w, "Error parsing time.", http.StatusInternalServerError)
				return
			}
			post.Datetime = datetime
		}

		post.Draft = len(r.Form["draft"]) == 1 && r.Form["draft"][0] == "on"

		err = post.Save(r.Context())
		if err != nil {
			log.Printf("err: %+v", err)
		}

		http.Redirect(w, r, "/admin/", http.StatusFound)
	})

	return r
}
