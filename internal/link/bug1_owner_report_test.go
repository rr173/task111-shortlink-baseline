package link

import (
	"context"
	"testing"
)

func TestOwnerReportNormalizesOwnerIdentity(t *testing.T) {
	svc := newSvc(t)
	if _, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://example.com", Owner: "alice"}); err != nil {
		t.Fatal(err)
	}
	report, err := svc.OwnerReport(context.Background(), " alice ")
	if err != nil {
		t.Fatal(err)
	}
	if report.Links != 1 || report.ActiveLinks != 1 {
		t.Fatalf("owner report = %+v, want one normalized link", report)
	}
}
