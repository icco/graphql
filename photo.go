package graphql

import (
	"context"
	"fmt"
	"io"
	"mime"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

const (
	// StorageBucketName is the bucket name we are uploading to.
	StorageBucketName = "icco-cloud"
)

func configureStorage(ctx context.Context, bucketID string) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

// Photo represents an uploaded photo
type Photo struct {
	ID          string `json:"id"`
	User        User   `json:"user"`
	Year        int
	ContentType string
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// Upload save the photo to GCS, and also makes sure the record is saved to the
// database.
func (p *Photo) Upload(ctx context.Context, f io.Reader) error {
	err := p.Save(ctx)
	if err != nil {
		return err
	}

	stgClient, err := configureStorage(ctx, StorageBucketName)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	uploader := stgClient.Object(p.Path()).NewWriter(tctx)
	uploader.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	uploader.ContentType = p.ContentType
	uploader.CacheControl = "public, max-age=86400"

	// Add a checksum.
	//uploader.CRC32C = crc32.Checksum(file, crc32.MakeTable(crc32.Castagnoli))
	//uploader.SendCRC32C = true

	_, err = io.Copy(uploader, f)
	if err != nil {
		return err
	}

	if err := uploader.Close(); err != nil {
		return err
	}

	return nil
}

// Save adds the photo to the database and checks that no data is missing.
func (p *Photo) Save(ctx context.Context) error {
	if p.ID == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		p.ID = uuid.String()
	}

	if p.Year == 0 {
		p.Year = time.Now().Year()
	}

	if p.Created.IsZero() {
		p.Created = time.Now()
	}

	p.Modified = time.Now()

	if _, err := db.ExecContext(
		ctx,
		`
INSERT INTO photos(id, year, content_type, user_id, created_at, modified_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET (year, content_type, user_id, created_at, modified_at) = ($2, $3, $4, $5, $6)
WHERE photos.id = $1;
`,
		p.ID,
		p.Year,
		p.ContentType,
		p.User.ID,
		p.Created,
		p.Modified); err != nil {
		return err
	}

	return nil
}

// Path returns the path the photo should be saved to.
func (p *Photo) Path() string {
	exts, err := mime.ExtensionsByType(p.ContentType)
	ext := ""
	if len(exts) > 0 && err == nil {
		ext = exts[0]
	} else {
		log.WithError(err).Warn("couldn't get an extension")
	}

	return fmt.Sprintf("photos/%d/%s%s", p.Year, p.ID, ext)
}

// URI returns the URI for this photo.
func (p *Photo) URI() string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", StorageBucketName, p.Path())
}
