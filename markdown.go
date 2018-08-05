package graphql

import (
	"html/template"
	"regexp"
	"strings"

	"gopkg.in/russross/blackfriday.v2"
)

var HashtagRegex *regexp.Regexp = regexp.MustCompile(`(\s)#(\w+)`)
var TwitterHandleRegex *regexp.Regexp = regexp.MustCompile(`(\s)@([_A-Za-z0-9]+)`)

// Markdown generator.
func Markdown(str string) template.HTML {
	inc := []byte(str)
	inc = twitterHandleToMarkdown(inc)
	inc = hashTagsToMarkdown(inc)
	s := blackfriday.Run(inc)
	return template.HTML(s)
}

// Takes a chunk of markdown and just returns the first paragraph.
func SummarizeText(str string) string {
	out := strings.Split(str, "\n")
	return out[0]
}

func twitterHandleToMarkdown(in []byte) []byte {
	return TwitterHandleRegex.ReplaceAll(in, []byte("$1[@$2](http://twitter.com/$2)"))
}

func hashTagsToMarkdown(in []byte) []byte {
	return HashtagRegex.ReplaceAll(in, []byte("$1[#$2](/tags/$2)"))
}
