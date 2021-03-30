package main

import (
	"net/http"

	"github.com/icco/graphql"
	"go.uber.org/zap"
)

func photoUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := graphql.GetUserFromContext(r.Context())
	if u == nil {
		err := Renderer.JSON(w, http.StatusForbidden, map[string]string{
			"error": "403: you must be logged in",
		})
		if err != nil {
			log.Errorw("could not render json", zap.Error(err))
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err == http.ErrMissingFile {
		err := Renderer.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "400: you must send a file",
		})
		if err != nil {
			log.Errorw("could not render json", zap.Error(err))
		}
		return
	} else if err != nil {
		log.Errorw("error reading file upload", zap.Error(err))
		internalErrorHandler(w, r)
		return
	}
	defer file.Close()

	p := &graphql.Photo{
		ContentType: header.Header.Get("Content-Type"),
		User:        *u,
	}

	err = p.Upload(ctx, file)
	if err != nil {
		log.Error("could not save image", zap.Error(err))
		internalErrorHandler(w, r)
		return
	}

	f := p.URI()
	err = Renderer.JSON(w, http.StatusOK, map[string]string{
		"upload": "ok",
		"file":   f.String(),
	})
	if err != nil {
		log.Error("could not render json", zap.Error(err))
	}
}
