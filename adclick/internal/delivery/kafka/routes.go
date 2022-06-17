package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adclick/internal/domain/entity"
	"github.com/crxfoz/teaserad/adclick/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type ClickService interface {
	AddBanner(ctx context.Context, url *entity.BannerURL) error
}

type Handler struct {
	clickSvc ClickService
}

func New(clickSvc ClickService) *Handler {
	return &Handler{clickSvc: clickSvc}
}

const (
	tracerName = "kafka-delivery"
)

func (h *Handler) OnNewBanner(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnNewBanner")
	defer span.End()

	var newBanner events.NewBanner
	if err := json.Unmarshal(msg.Value, &newBanner); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	if err := h.clickSvc.AddBanner(spanCtx, &entity.BannerURL{
		BannerID:  newBanner.BannerID,
		BannerURL: newBanner.BannerURL,
	}); err != nil {
		return fmt.Errorf("could not add banner: %w", err)
	}

	return nil
}
