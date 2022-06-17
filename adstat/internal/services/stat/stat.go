package stat

import (
	"context"
	"fmt"
	"time"

	"github.com/crxfoz/teaserad/adstat/internal/domain"
	"github.com/crxfoz/teaserad/adstat/internal/domain/entity"
)

type StatRepo interface {
	GetBannerStat(ctx context.Context, bannerID int, from time.Time) ([]*entity.BannerStat, error)
	GetPlatformStat(ctx context.Context, platformID int, from time.Time) ([]*entity.PlatformStat, error)
}

type StatService struct {
	repo   StatRepo
	logger domain.Logger
}

func New(repo StatRepo, logger domain.Logger) *StatService {
	return &StatService{repo: repo, logger: logger}
}

func (s *StatService) GetBannerStat(ctx context.Context, bannerID int, from time.Time) ([]*entity.BannerStat, error) {
	stat, err := s.repo.GetBannerStat(ctx, bannerID, from)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(stat) == 0 {
		return []*entity.BannerStat{}, nil
	}

	return stat, nil
}

func (s *StatService) GetBannerStatToday(ctx context.Context, bannerID int) ([]*entity.BannerStat, error) {
	stat, err := s.repo.GetBannerStat(ctx, bannerID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(stat) == 0 {
		return []*entity.BannerStat{}, nil
	}

	return stat, nil
}

func (s *StatService) GetPlatformStat(ctx context.Context, platformID int, from time.Time) ([]*entity.PlatformStat, error) {
	stat, err := s.repo.GetPlatformStat(ctx, platformID, from)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(stat) == 0 {
		return []*entity.PlatformStat{}, nil
	}

	return stat, nil
}

func (s *StatService) GetPlatformStatToday(ctx context.Context, platformID int) ([]*entity.PlatformStat, error) {
	stat, err := s.repo.GetPlatformStat(ctx, platformID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(stat) == 0 {
		return []*entity.PlatformStat{}, nil
	}

	return stat, nil
}
