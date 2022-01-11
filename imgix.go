package graphql

import (
	"os"

	ix "github.com/imgix/imgix-go/v2"
)

func GenerateSocialImage() {
	ixToken := os.Getenv("IX_TOKEN")
	ix.NewURLBuilder("demo.imgix.net", ix.WithToken(ixToken), ix.WithLibParam(false))
}
