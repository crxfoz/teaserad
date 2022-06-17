package banner

import (
	"context"
	"fmt"

	"github.com/crxfoz/teaserad/adeliver/internal/domain/entity"
	"github.com/crxfoz/teaserad/adeliver/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type BannerRepo interface {
	GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error)
	AddBanner(ctx context.Context, banner entity.Banner) error
	GetClick(ctx context.Context, bannerID int) (int64, error)
	GetShows(ctx context.Context, bannerID int) (int64, error)
	GetSpend(ctx context.Context, bannerID int) (float64, error)
	AddClick(ctx context.Context, bannerID int) (int64, error)
	AddShow(ctx context.Context, bannerID int) (int64, error)
	AddSpend(ctx context.Context, bannerID int, price float64) (float64, error)
}

// BannerNotify signal other services that banner has been stopped because reached its limits
type BannerNotify interface {
	NotifyBannerStopped(context.Context, events.BannerReachedLimits) error
}

type BannerDispatcher interface {
	StartBanner(ctx context.Context, started events.BannerStart) error
	StopBanner(ctx context.Context, stopped events.BannerStop) error
}

type BannerService struct {
	repo       BannerRepo
	notify     BannerNotify
	dispatcher BannerDispatcher
}

func New(repo BannerRepo, notify BannerNotify, dispatcher BannerDispatcher) *BannerService {
	return &BannerService{repo: repo, notify: notify, dispatcher: dispatcher}
}

const (
	tracerName = "usecase"
)

func (b *BannerService) StopBanner(ctx context.Context, incoming events.BannerStoppedIncoming) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StopBanner")
	defer span.End()

	if err := b.dispatcher.StopBanner(spanCtx, events.BannerStop{BannerID: incoming.BannerID}); err != nil {
		return fmt.Errorf("could not stop banner on dispatcher: %w", err)
	}
	return nil
}

func (b *BannerService) StartBanner(ctx context.Context, incoming events.BannerStartedIncoming) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StartBanner")
	defer span.End()

	_, err := b.repo.GetBanner(spanCtx, incoming.BannerID)
	if err != nil {
		errCreate := b.repo.AddBanner(spanCtx, entity.Banner{
			ID:          incoming.BannerID,
			LimitShows:  incoming.LimitShows,
			LimitClicks: incoming.LimitClicks,
			LimitBudget: incoming.LimitBudget,
		})
		if errCreate != nil {
			return fmt.Errorf("could not add new banner: %w", err)
		}
	}

	err = b.dispatcher.StartBanner(spanCtx, events.BannerStart{
		BannerID:    incoming.BannerID,
		UserID:      incoming.UserID,
		ImgData:     incoming.ImgData,
		BannerText:  incoming.BannerText,
		BannerURL:   incoming.BannerURL,
		LimitShows:  incoming.LimitShows,
		LimitClicks: incoming.LimitClicks,
		LimitBudget: incoming.LimitBudget,
		Device:      incoming.Device,
		CategoryID:  incoming.CategoryID,
	})

	if err != nil {
		return fmt.Errorf("could not start banner on dispatcher: %w", err)
	}

	return nil
}

func (b *BannerService) NewClick(ctx context.Context, incoming events.Click) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "NewClick")
	defer span.End()

	limits, err := b.repo.GetBanner(spanCtx, incoming.BannerID)
	if err != nil {
		return fmt.Errorf("could not get banner: %w", err)
	}

	clicks, err := b.repo.AddClick(spanCtx, incoming.BannerID)
	if err != nil {
		return fmt.Errorf("could not add click: %w", err)
	}

	if clicks > limits.LimitClicks {
		err = b.notify.NotifyBannerStopped(spanCtx, events.BannerReachedLimits{
			BannerID: incoming.BannerID,
			Reason:   "clicks"})
		if err != nil {
			return fmt.Errorf("could not notify banner to stop: %w", err)
		}

		err = b.dispatcher.StopBanner(spanCtx, events.BannerStop{BannerID: incoming.BannerID})
		if err != nil {
			return fmt.Errorf("could not stop banner on dispatcher: %w", err)
		}
	}

	return nil
}

func (b *BannerService) NewView(ctx context.Context, incoming events.View) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "NewView")
	defer span.End()

	limits, err := b.repo.GetBanner(spanCtx, incoming.BannerID)
	if err != nil {
		return fmt.Errorf("could not get banner: %w", err)
	}

	views, err := b.repo.AddShow(spanCtx, incoming.BannerID)
	if err != nil {
		return fmt.Errorf("could not add click: %w", err)
	}

	if views > limits.LimitShows {
		err = b.notify.NotifyBannerStopped(spanCtx, events.BannerReachedLimits{
			BannerID: incoming.BannerID,
			Reason:   "views"})
		if err != nil {
			return fmt.Errorf("could not notify banner to stop: %w", err)
		}

		err = b.dispatcher.StopBanner(spanCtx, events.BannerStop{BannerID: incoming.BannerID})
		if err != nil {
			return fmt.Errorf("could not stop banner on dispatcher: %w", err)
		}
	}

	return nil
}
