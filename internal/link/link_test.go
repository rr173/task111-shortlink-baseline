package link

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"task111-shortlink/internal/store"
)

func newSvc(t *testing.T) *Service {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "l.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return New(s)
}

func TestCreateResolve(t *testing.T) {
	svc := newSvc(t)
	l, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://x.com"})
	if err != nil {
		t.Fatal(err)
	}
	if l.Code == "" {
		t.Fatal("empty code")
	}
	got, err := svc.Resolve(context.Background(), l.Code)
	if err != nil {
		t.Fatal(err)
	}
	if got.TargetURL != "https://x.com" {
		t.Fatal("resolve target mismatch")
	}
}

func TestResolveNotFound(t *testing.T) {
	svc := newSvc(t)
	_, err := svc.Resolve(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestCreateInvalidURL(t *testing.T) {
	svc := newSvc(t)
	_, err := svc.Create(context.Background(), CreateReq{TargetURL: "not-a-url"})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("want ErrInvalidURL, got %v", err)
	}
}

func TestCustomCodeUnique(t *testing.T) {
	svc := newSvc(t)
	if _, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://x.com", CustomCode: "my"}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://y.com", CustomCode: "my"})
	if err == nil {
		t.Fatal("expected duplicate code error")
	}
}

func TestUpdateDelete(t *testing.T) {
	svc := newSvc(t)
	l, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://x.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Update(context.Background(), l.Code, "https://y.com", "desc", 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(context.Background(), l.Code); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Resolve(context.Background(), l.Code); !errors.Is(err, ErrNotFound) {
		t.Fatal("link should be gone")
	}
}

func TestStatusNoLimit(t *testing.T) {
	svc := newSvc(t)
	l, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://x.com"})
	if err != nil {
		t.Fatal(err)
	}
	st, err := svc.Status(context.Background(), l.Code)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Active {
		t.Fatal("should be active")
	}
	if st.Remaining != -1 {
		t.Fatalf("remaining = %d, want -1", st.Remaining)
	}
}

// TestOwnerReportTrimsOwner 验证按负责人查看报表时，输入名称带首尾空格
// 仍能正确归属统计：创建路径会对负责人做 TrimSpace 后持久化，因此报表
// 查询键也必须先归一化，否则精确匹配不到任何行，链接数与活跃数会误判为 0。
func TestOwnerReportTrimsOwner(t *testing.T) {
	svc := newSvc(t)
	if _, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://a.com", Owner: "  alice  "}); err != nil {
		t.Fatal(err)
	}
	rep, err := svc.OwnerReport(context.Background(), "  alice  ")
	if err != nil {
		t.Fatal(err)
	}
	if rep.Links != 1 {
		t.Fatalf("links = %d, want 1 (leading/trailing whitespace must not break attribution)", rep.Links)
	}
	if rep.ActiveLinks != 1 {
		t.Fatalf("active_links = %d, want 1", rep.ActiveLinks)
	}
	if rep.Owner != "alice" {
		t.Fatalf("owner = %q, want normalized \"alice\"", rep.Owner)
	}
}

// TestOwnerReportIsolatesOwners 验证不同负责人的数据在归一化后仍彼此隔离：
// 相同负责人带不同空格应归到同一报表，且不应把其他负责人的链接计入。
func TestOwnerReportIsolatesOwners(t *testing.T) {
	svc := newSvc(t)
	if _, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://a.com", Owner: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(context.Background(), CreateReq{TargetURL: "https://b.com", Owner: "bob"}); err != nil {
		t.Fatal(err)
	}
	alice, err := svc.OwnerReport(context.Background(), "  alice ")
	if err != nil {
		t.Fatal(err)
	}
	if alice.Links != 1 || alice.Owner != "alice" {
		t.Fatalf("alice report = %+v, want links=1 owner=alice", alice)
	}
	bob, err := svc.OwnerReport(context.Background(), "bob")
	if err != nil {
		t.Fatal(err)
	}
	if bob.Links != 1 || bob.Owner != "bob" {
		t.Fatalf("bob report = %+v, want links=1 owner=bob", bob)
	}
}
