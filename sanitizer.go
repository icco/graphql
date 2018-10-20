package graphql

import (
	"fmt"
	"strings"
)

// sanitize will replace existing backticks with escape chars, then wrap in backticks.
func sanitize(str string) string {
	withEscapedBackticks := strings.Replace(str, "`", "\\`", -1)
	wrappedSanitizedString := fmt.Sprintf("`%s`", withEscapedBackticks)
	return wrappedSanitizedString;
}