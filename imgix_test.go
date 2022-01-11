package graphql

import (
	"context"
	"testing"
	"time"
)

func TestGenerateSocialImage(t *testing.T) {
	ctx := context.TODO()

	u, err := GenerateSocialImage(ctx, "Test", time.Now())
	if err != nil {
		t.Error(err)
	}

	t.Logf("Generate log %q", u)
}
