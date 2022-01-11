package graphql

import (
	"context"
	"time"

	ix "github.com/imgix/imgix-go/v2"
)

// GenerateSocialImage creates a static image URL for a post.
func GenerateSocialImage(ctx context.Context, title string, when time.Time) (*URI, error) {
	bgColor := "eeeceb"
	txtColor := "333"
	txtFont := []string{"Roboto Slab", "serif"}

	ub := ix.NewURLBuilder("icco.imgix.net")

	formatedTitle := ub.CreateURL("/~text",
		ix.Param("bg", bgColor),
		ix.Param("h", "260"),
		ix.Param("txt64", title),
		ix.Param("txtalign", "bottom"),
		ix.Param("txtclr", txtColor),
		ix.Param("txtfont64", txtFont...),
		ix.Param("txtsize", "72"),
		ix.Param("w", "1100"),
	)

	urlString := ub.CreateURL("/canvas.png", []ix.IxParam{
		ix.Param("ba", "bottom"),
		ix.Param("bg", bgColor),
		ix.Param("blend64", formatedTitle),
		ix.Param("bm", "normal"),
		ix.Param("bx", "50"),
		ix.Param("by", "280"),
		ix.Param("fit", "crop"),
		ix.Param("fm", "png8"),
		ix.Param("h", "630"),
		ix.Param("mark-w", "200"),
		ix.Param("mark64", "https://natwelch.com/i/logo.png"),
		ix.Param("markx", "20"),
		ix.Param("marky", "28"),
		ix.Param("txt64", when.Format("2006-01-02")),
		ix.Param("txtalign", "left", "bottom"),
		ix.Param("txtclip", "end", "ellipsis"),
		ix.Param("txtclr", txtColor),
		ix.Param("txtfont64", txtFont...),
		ix.Param("txtpad", "60"),
		ix.Param("txtsize", "24"),
		ix.Param("w", "1200"),
	}...)

	return NewURI(urlString), nil
}
