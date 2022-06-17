package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adeliver/internal/domain/events"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
)

const (
	topicReachedLimits = "adeliver.banner.limits"
)

type BannerRepo struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *BannerRepo {
	return &BannerRepo{conn: conn}
}

const (
	tracerName = "kafka-producer"
)

func (r *BannerRepo) NotifyBannerStopped(ctx context.Context, event events.BannerReachedLimits) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "NotifyBannerStopped")
	defer span.End()

	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicReachedLimits,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(spanCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = r.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
