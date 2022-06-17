package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adeliver/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type BannerService interface {
	StopBanner(ctx context.Context, incoming events.BannerStoppedIncoming) error
	StartBanner(ctx context.Context, incoming events.BannerStartedIncoming) error
	NewClick(ctx context.Context, incoming events.Click) error
	NewView(ctx context.Context, incoming events.View) error
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

	var stopped events.BannerStoppedIncoming
	if err := json.Unmarshal(msg.Value, &stopped); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.StopBanner(spanCtx, stopped)
}

func (c *Consumer) OnBannerStarted(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnBannerStarted")
	defer span.End()

	var started events.BannerStartedIncoming
	if err := json.Unmarshal(msg.Value, &started); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.StartBanner(spanCtx, started)
}

func (c *Consumer) OnActionView(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnActionView")
	defer span.End()

	var view events.View
	if err := json.Unmarshal(msg.Value, &view); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.NewView(spanCtx, view)
}

func (c *Consumer) OnActionClick(ctx context.Context, msg *sarama.ConsumerMessage) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "OnActionClick")
	defer span.End()

	var click events.Click
	if err := json.Unmarshal(msg.Value, &click); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return c.bannerSvc.NewClick(spanCtx, click)
}
