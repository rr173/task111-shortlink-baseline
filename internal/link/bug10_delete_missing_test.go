package link

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteMissingLinkReturnsNotFound(t *testing.T) {
	svc := newSvc(t)
	if err := svc.Delete(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete missing error = %v, want ErrNotFound", err)
	}
}
