package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/crmad/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type BannerService interface {
	BannerUpdated(ctx context.Context, updated events.BannerUpdated) error
	BannerReachedLimits(ctx context.Context, item events.BannerReachedLimits) error
}

type BannerStatus struct {
	bannerSvc BannerService
}

func New(bannerSvc BannerService) *BannerStatus {
	return &BannerStatus{bannerSvc: bannerSvc}
}

const (
	tracerName = "kafka-delivery"
)

func (bs *BannerStatus) OnBannerUpdated(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnBannerUpdated")
	defer span.End()

	var updated events.BannerUpdated
	if err := json.Unmarshal(msg.Value, &updated); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return bs.bannerSvc.BannerUpdated(spanCtx, updated)
}

func (bs *BannerStatus) OnBannerReachedLimits(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnBannerReachedLimits")
	defer span.End()

	var reached events.BannerReachedLimits
	if err := json.Unmarshal(msg.Value, &reached); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return bs.bannerSvc.BannerReachedLimits(spanCtx, reached)
}
