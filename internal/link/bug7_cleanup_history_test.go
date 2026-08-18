package link

import (
	"context"
	"testing"
	"time"

	"task111-shortlink/internal/store"
)

func TestCleanupExpiredKeepsClickHistory(t *testing.T) {
	svc := newSvc(t)
	l, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://example.com", ExpiresAt: time.Now().Add(-time.Minute).UnixMilli()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.store.InsertClick(context.Background(), store.Click{Code: l.Code, Fingerprint: "fp"}); err != nil {
		t.Fatal(err)
	}
	report, err := svc.CleanupExpired(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if report.Deleted != 1 {
		t.Fatalf("cleanup report = %+v", report)
	}
	n, err := svc.store.CountClicks(context.Background(), l.Code)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("click history count = %d, want 1", n)
	}
}
