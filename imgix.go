package graphql

import (
	"os"

	ix "github.com/imgix/imgix-go/v2"
)

// https://assets.imgix.net/page-weight/canvas.png?bg=eeeceb&mark64=aHR0cHM6Ly9uYXR3ZWxjaC5jb20vaS9sb2dvLnBuZz9oPTEwMA&w=1200&h=630&fit=crop&bm=normal&bx=50&ba=bottom&markx=40&txt64=MjAyMi0wMS0xMA&txtalign=left%2Cbottom&txtsize=24&txtclr=000000&txtpad=60&txtclip=end%2Cellipsis&by=280&txtfont64=RGluIEFsdGVybmF0ZQ&fm=png8&marky=28&mark-w=200
func GenerateSocialImage() {
	ixToken := os.Getenv("IX_TOKEN")
	ix.NewURLBuilder("demo.imgix.net", ix.WithToken(ixToken), ix.WithLibParam(false))
}
