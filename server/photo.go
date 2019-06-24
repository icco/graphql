package main

import (
	"net/http"

	"github.com/icco/graphql"
)

func photoUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := graphql.GetUserFromContext(r.Context())
	if u == nil {
		err := Renderer.JSON(w, http.StatusForbidden, map[string]string{
			"error": "403: you must be logged in",
		})
		if err != nil {
			log.WithError(err).Error("could not render json")
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err == http.ErrMissingFile {
		err := Renderer.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "400: you must send a file",
		})
		if err != nil {
			log.WithError(err).Error("could not render json")
		}
		return
	} else if err != nil {
		log.WithError(err).Error("error reading file upload")
		internalErrorHandler(w, r)
		return
	}
	defer file.Close()

	log.WithField("file_header", header).Debug("received file")

	p := &graphql.Photo{
		ContentType: header.Header.Get("Content-Type"),
		User:        *u,
	}

	err = p.Upload(ctx, file)
	if err != nil {
		log.WithError(err).Error("could not save image")
		internalErrorHandler(w, r)
		return
	}

	f := p.URI()
	err = Renderer.JSON(w, http.StatusOK, map[string]string{
		"upload": "ok",
		"file":   f.String(),
	})
	if err != nil {
		log.WithError(err).Error("could not render json")
	}
}
