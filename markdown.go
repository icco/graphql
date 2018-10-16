package graphql

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

var (
	// HashtagRegex is a regex for finding hashtags in Markdown.
	HashtagRegex = regexp.MustCompile(`(\s)#(\w+)`)

	// TwitterHandleRegex is a regex for finding @username in Markdown.
	TwitterHandleRegex = regexp.MustCompile(`(\s)@([_A-Za-z0-9]+)`)
)

// Markdown generator.
func Markdown(str string) template.HTML {
	inc := []byte(str)
	inc = twitterHandleToMarkdown(inc)
	inc = hashTagsToMarkdown(inc)
	s := blackfriday.Run(inc)
	// TODO: https://github.com/microcosm-cc/bluemonday
	return template.HTML(s)
}

// SummarizeText takes a chunk of markdown and just returns the first paragraph.
func SummarizeText(str string) string {
	out := strings.Split(str, "\n")
	return strings.TrimSpace(out[0])
}

func twitterHandleToMarkdown(in []byte) []byte {
	return TwitterHandleRegex.ReplaceAll(in, []byte("$1[@$2](http://twitter.com/$2)"))
}

func hashTagsToMarkdown(in []byte) []byte {
	return HashtagRegex.ReplaceAll(in, []byte("$1[#$2](/tags/$2)"))
}
