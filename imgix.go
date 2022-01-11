package graphql

import (
	"context"
	"time"

	ix "github.com/imgix/imgix-go/v2"
)

// GenerateSocialImage creates a static image URL for a post.
func GenerateSocialImage(ctx context.Context, title string, when time.Time) (*URI, error) {
	ub := ix.NewURLBuilder("icco.imgix.net")
	urlString := ub.CreateURL("/canvas.png", []ix.IxParam{
		ix.Param("bg", "eeeceb"),
		ix.Param("mark64", "aHR0cHM6Ly9uYXR3ZWxjaC5jb20vaS9sb2dvLnBuZz9oPTEwMA"),
		ix.Param("w", "1200"),
		ix.Param("h", "630"),
		ix.Param("fit", "crop"),
		ix.Param("bm", "normal"),
		ix.Param("bx", "50"),
		ix.Param("ba", "bottom"),
		ix.Param("markx", "40"),
		ix.Param("txt64", "MjAyMi0wMS0xMA"),
		ix.Param("txtalign", "left", "bottom"),
		ix.Param("txtsize", "24"),
		ix.Param("txtclr", "000000"),
		ix.Param("txtpad", "60"),
		ix.Param("txtclip", "end", "ellipsis"),
		ix.Param("by", "280"),
		ix.Param("txtfont64", "RGluIEFsdGVybmF0ZQ"),
		ix.Param("fm", "png8"),
		ix.Param("marky", "28"),
		ix.Param("mark-w", "200"),
	}...)

	return NewURI(urlString), nil
}
