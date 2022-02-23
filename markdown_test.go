package graphql

import (
	"strings"
	"testing"
)

func TestMarkdown(*testing.T) {
	tests := map[string]struct {
		test     string
		contains string
	}{
		"twitter": {
			test:     "@icco",
			contains: "https://twitter.com/icco",
		},
	}

	for i, tc := range tests {
		tc := tc // capture range variable
		t.Run(i, func(t *testing.T) {
			t.Parallel()
			if got := Markdown(tc.test); !strings.Contains(got, tc.contains) {
				t.Errorf("expected Markdown(%q) to contain %q. got %q", tc.test, tc.contains, got)
			}
		})
	}
}
