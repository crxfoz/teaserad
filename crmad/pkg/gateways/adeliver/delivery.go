package adeliver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/crmad/internal/domain/events"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
)

const (
	topicBanenrStart = "adeliver.banner.start"
	topicBannerStop  = "adeliver.banner.stop"
)

type Producer struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *Producer {
	return &Producer{conn: conn}
}

const (
	tracerName = "kafka-producer-adelivery"
)

func (b *Producer) BannerStart(ctx context.Context, msg events.BannerStart) error {
	newCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerStart")
	defer span.End()

	out, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicBanenrStart,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(newCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = b.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func (b *Producer) BannerStop(ctx context.Context, msg events.BannerStop) error {
	newCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerStart")
	defer span.End()

	out, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicBannerStop,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(newCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = b.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
