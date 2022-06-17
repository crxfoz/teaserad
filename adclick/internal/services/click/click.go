package click

import (
	"context"
	"fmt"
	"time"

	"github.com/crxfoz/teaserad/adclick/internal/domain"
	"github.com/crxfoz/teaserad/adclick/internal/domain/entity"
	"github.com/crxfoz/teaserad/adclick/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type ClickRepo interface {
	AddBanner(ctx context.Context, banner *entity.BannerURL) error
	GetBanner(ctx context.Context, bannerID int) (*entity.BannerURL, error)
}

type ClickNotifier interface {
	SendClick(ctx context.Context, event *events.Click) error
}

type Service struct {
	clickRepo     ClickRepo
	clickNotifier ClickNotifier
	logger        domain.Logger
}

func New(clickRepo ClickRepo, clickNotifier ClickNotifier, logger domain.Logger) *Service {
	return &Service{clickRepo: clickRepo, clickNotifier: clickNotifier, logger: logger}
}

const (
	tracerName = "usecase"
)

func (s *Service) AddBanner(ctx context.Context, banner *entity.BannerURL) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddBanner")
	defer span.End()

	if err := s.clickRepo.AddBanner(spanCtx, banner); err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (s *Service) NewClick(ctx context.Context, bannerID int, platformID int, viewID string) (*entity.BannerURL, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "NewClick")
	defer span.End()

	info, err := s.clickRepo.GetBanner(spanCtx, bannerID)
	if err != nil {
		return nil, fmt.Errorf("could not get banner - %d: %w", bannerID, err)
	}

	err = s.clickNotifier.SendClick(spanCtx, &events.Click{
		BannerID:   bannerID,
		PlatformID: platformID,
		ViewID:     viewID,
		Price:      1,
		CreatedAt:  time.Now().UTC().Unix(),
	})
	if err != nil {
		return nil, fmt.Errorf("could not send click: %w", err)
	}

	return info, nil
}
