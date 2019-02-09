package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/icco/graphql"
)

var (
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

func init() {
	var err error
	StorageBucketName = "icco-cloud"
	StorageBucket, err = configureStorage(StorageBucketName)

	if err != nil {
		log.Fatal(err)
	}
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

func photoUploadHandler(w http.ResponseWriter, r *http.Request) {
	u := graphql.ForContext(r.Context())
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

	// Example: {"Content-Disposition":["form-data; name=\"file\"; filename=\"test.jpg\""],"Content-Type":["image/jpeg"]},"Size":3422342}
	log.WithField("file_header", header).Debug("recieved file")
	id, err := uuid.NewRandom()
	if err != nil {
		log.WithError(err).Error("error generating random")
		internalErrorHandler(w, r)
		return
	}
	name := fmt.Sprintf("photos/%d/%s%s", time.Now().Year(), id, path.Ext(header.Filename))
	uploader := StorageBucket.Object(name).NewWriter(r.Context())
	uploader.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	uploader.ContentType = header.Header.Get("Content-Type")
	uploader.CacheControl = "public, max-age=86400"
	_, err = io.Copy(uploader, file)
	if err != nil {
		log.WithError(err).Error("error reading file upload")
		internalErrorHandler(w, r)
		return
	}

	err = Renderer.JSON(w, http.StatusOK, map[string]string{
		"upload": "ok",
		"file":   fmt.Sprintf("https://storage.googleapis.com/%s/%s", StorageBucketName, name),
	})
	if err != nil {
		log.WithError(err).Error("could not render json")
	}
}
