package graphql

import (
	"context"
	"fmt"
	"io"
	"mime"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

// IsLinkable exists to show that this method implements the Linkable type in
// graphql.
func (p *Photo) IsLinkable() {}

// Upload save the photo to GCS, and also makes sure the record is saved to the
// database.
func (p *Photo) Upload(ctx context.Context, f io.Reader) error {
	if err := p.Save(ctx); err != nil {
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

	if _, err := io.Copy(uploader, f); err != nil {
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
		log.Warnw("couldn't get an extension", zap.Error(err))
	}

	return fmt.Sprintf("photos/%d/%s%s", p.Year, p.ID, ext)
}

// URI returns the URI for this photo.
func (p *Photo) URI() *URI {
	return NewURI(fmt.Sprintf("https://icco.imgix.net/%s", p.Path()))
}

func (p *Photo) GetURI() URI {
	return *p.URI()
}

// UserPhotos gets all photos for a User.
func UserPhotos(ctx context.Context, u *User, limit int, offset int) ([]*Photo, error) {
	if u == nil {
		return nil, fmt.Errorf("no user specified")
	}

	rows, err := db.QueryContext(
		ctx, `
    SELECT id, year, content_type, user_id, created_at, modified_at
    FROM photos
    WHERE user_id = $1
    ORDER BY datetime DESC
    LIMIT $2 OFFSET $3
    `,
		u.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := make([]*Photo, 0)
	for rows.Next() {
		p := &Photo{}
		err := rows.Scan(
			&p.ID,
			&p.Year,
			&p.ContentType,
			&p.User.ID,
			&p.Created,
			&p.Modified,
		)
		if err != nil {
			return nil, err
		}

		photos = append(photos, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return photos, nil
}
