package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type BannerService interface {
	NewBanner(ctx context.Context, banner events.BannerCreated) error
}

type BannerConsumer struct {
	bannerSvc BannerService
}

func New(bannerSvc BannerService) *BannerConsumer {
	return &BannerConsumer{bannerSvc: bannerSvc}
}

func (bc *BannerConsumer) OnNewBanner(ctx context.Context, msg *sarama.ConsumerMessage) error {
	newCtx, span := otel.Tracer("kafka-consumer").Start(ctx, "OnNewBanner")
	defer span.End()

	var created events.BannerCreated
	if err := json.Unmarshal(msg.Value, &created); err != nil {
		return fmt.Errorf("could not parse message: %w", err)
	}

	return bc.bannerSvc.NewBanner(newCtx, created)
}
