package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adshow/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type BannerService interface {
	StartBanner(ctx context.Context, banner *events.BannerStart) error
	StopBanner(ctx context.Context, bannerID int) error
}

type Consumer struct {
	bannerSvc BannerService
}

func New(bannerSvc BannerService) *Consumer {
	return &Consumer{bannerSvc: bannerSvc}
}

const (
	tracerName = "kafka-delivery"
)

func (c *Consumer) OnBannerStopped(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnBannerStopped")
	defer span.End()

	var stopped events.BannerStop
	if err := json.Unmarshal(msg.Value, &stopped); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.StopBanner(spanCtx, stopped.BannerID)
}

func (c *Consumer) OnBannerStarted(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnBannerStarted")
	defer span.End()

	var started events.BannerStart
	if err := json.Unmarshal(msg.Value, &started); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.StartBanner(spanCtx, &started)
}
