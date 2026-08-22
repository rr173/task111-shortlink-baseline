package link

import "context"

// OwnerReport is a persisted owner-level view used by quota and cleanup jobs.
type OwnerReport struct {
	Owner       string `json:"owner"`
	Links       int    `json:"links"`
	Clicks      int    `json:"clicks"`
	ActiveLinks int    `json:"active_links"`
}

func (s *Service) OwnerReport(ctx context.Context, owner string) (OwnerReport, error) {
	// 与创建路径一致地归一化负责人：创建时已通过 NormalizeOwner 去除
	// 首尾空格后持久化，因此查询时若带空格会精确匹配不到任何行，
	// 导致本应属于该负责人的链接数与活跃数变为零。归一化查询键后，
	// 首尾空格不再改变统计归属，且不同负责人仍归一化为不同键，数据保持隔离。
	owner = NormalizeOwner(owner)
	summary, err := s.store.OwnerSummary(ctx, owner)
	if err != nil {
		return OwnerReport{}, err
	}
	return OwnerReport{Owner: owner, Links: summary.Links, Clicks: summary.Clicks, ActiveLinks: summary.ActiveLinks}, nil
}
