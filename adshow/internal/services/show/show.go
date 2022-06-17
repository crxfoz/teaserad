package show

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/crxfoz/teaserad/adshow/internal/domain/entity"
	"github.com/crxfoz/teaserad/adshow/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type ShowRepo interface {
	AddBanner(ctx context.Context, start *events.BannerStart, toPlatforms []int) error
	DeleteBannerAll(ctx context.Context, bannerID int) error
	BannersForPlatform(ctx context.Context, platformID int, deviceType string, limit int) ([]*entity.Banner, error)
}

type PlatformRepo interface {
	GetPlatformsByCategory(ctx context.Context, categoryID int) (entity.Platforms, error)
	GetPlatforms(ctx context.Context) (entity.Platforms, error)
	GetPlatform(ctx context.Context, platformID int) (*entity.Platform, error)
	AddPlatform(ctx context.Context, platform *entity.Platform) error
	DeletePlatform(ctx context.Context, platformID int) error
}

type ShowNotifier interface {
	AddViews(context.Context, []*events.View) error
}

type ShowService struct {
	repo         ShowRepo
	platformRepo PlatformRepo
	showNotifier ShowNotifier
}

const (
	tracerName = "usecase"
)

func New(repo ShowRepo, platformRepo PlatformRepo, showNotifier ShowNotifier) *ShowService {
	rand.Seed(time.Now().UnixNano())

	return &ShowService{repo: repo, platformRepo: platformRepo, showNotifier: showNotifier}
}

func (s *ShowService) StartBanner(ctx context.Context, banner *events.BannerStart) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StartBanner")
	defer span.End()

	platforms, err := s.platformRepo.GetPlatformsByCategory(ctx, banner.CategoryID)
	if err != nil {
		return fmt.Errorf("could not get platforms: %w", err)
	}

	platforms.Shuffle()

	if err := s.repo.AddBanner(spanCtx, banner, platforms.Ids()); err != nil {
		return fmt.Errorf("could not start banner: %w", err)
	}

	return nil
}

func (s *ShowService) StopBanner(ctx context.Context, bannerID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StopBanner")
	defer span.End()

	if err := s.repo.DeleteBannerAll(spanCtx, bannerID); err != nil {
		return fmt.Errorf("could not stop banner: %w", err)
	}

	return nil
}

func (s *ShowService) AddPlatform(ctx context.Context, platform *entity.Platform) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddPlatform")
	defer span.End()

	if err := s.platformRepo.AddPlatform(spanCtx, platform); err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (s *ShowService) GetBannersForPlatform(ctx context.Context, platformID int, deviceType string, limit int) ([]*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBannersForPlatform")
	defer span.End()

	banners, err := s.repo.BannersForPlatform(spanCtx, platformID, deviceType, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get banners: %w", err)
	}

	return banners, nil
}

func (s *ShowService) ShowBanners(ctx context.Context, limit int, hitCtx *entity.HitContext) ([]*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "ShowBanners")
	defer span.End()

	deviceType := hitCtx.DeviceType()

	banners, err := s.repo.BannersForPlatform(spanCtx, hitCtx.PlatformID, deviceType, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get banners: %w", err)
	}

	views := make([]*events.View, 0, len(banners))
	for _, item := range banners {
		views = append(views, &events.View{
			BannerID:   item.BannerID,
			PlatformID: hitCtx.PlatformID,
			UserAgent:  hitCtx.UserAgent,
			Device:     deviceType,
			CreatedAt:  time.Now().UTC().Unix(),
		})
	}

	if err := s.showNotifier.AddViews(spanCtx, views); err != nil {
		return nil, fmt.Errorf("could not register views: %w", err)
	}

	return banners, nil
}
