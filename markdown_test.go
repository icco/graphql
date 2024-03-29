package graphql

import (
	"strings"
	"testing"
)

func TestMarkdown(t *testing.T) {
	tests := map[string]struct {
		test     string
		contains string
	}{
		"twitter": {
			test:     " @icco",
			contains: "https://twitter.com/icco",
		},
		"list": {
			test: `
 - a
 - b
 - c
      `,
			contains: "li",
		},
		"hashtag": {
			test:     ` #happy`,
			contains: `/tags/happy`,
		},
	}

	for i, tc := range tests {
		tc := tc // capture range variable
		t.Run(i, func(t *testing.T) {
			t.Parallel()
			if got := Markdown(tc.test); !strings.Contains(string(got), tc.contains) {
				t.Errorf("expected Markdown(%q) to contain %q. got %q", tc.test, tc.contains, got)
			}
		})
	}
}
