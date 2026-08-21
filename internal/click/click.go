// Package click 负责访问点击的采集与去重。
package click

import (
	"context"
	"time"

	"task111-shortlink/internal/store"
)

// Service 聚合点击相关操作。
type Service struct {
	store *store.Store
}

// New 构造点击服务。
func New(s *store.Store) *Service { return &Service{store: s} }

// Record 写入一条点击记录，返回落库后的完整记录。
// fingerprint 用于同一浏览器在短时间内重复刷新时的去重。
func (s *Service) Record(ctx context.Context, code, referer, ua, ip, fingerprint string) (store.Click, error) {
	return s.store.InsertClick(ctx, store.Click{
		Code:        code,
		Referer:     referer,
		UserAgent:   ua,
		IP:          ip,
		Fingerprint: NormalizeFingerprint(fingerprint),
		ClickedAt:   time.Now().UnixMilli(),
	})
}

// Recent 返回最近的点击记录。
func (s *Service) Recent(ctx context.Context, limit int) ([]store.Click, error) {
	limit = NormalizeLimit(limit)
	return s.store.RecentClicks(ctx, limit)
}

// RecentByCode 返回指定短码最近的点击记录，时间范围与排序规则与 Recent
// 一致（按 clicked_at 倒序取最近 limit 条），仅把归属范围限定在该短码，
// 用于单短码活动窗口，避免混入其他短码的访问记录。
func (s *Service) RecentByCode(ctx context.Context, code string, limit int) ([]store.Click, error) {
	limit = NormalizeLimit(limit)
	return s.store.RecentClicksByCode(ctx, code, limit)
}
