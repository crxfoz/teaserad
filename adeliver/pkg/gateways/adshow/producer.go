package adshow

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
	topicStart = "adshow.banner.start"
	topicStop  = "adshow.banner.stop"
)

type Producer struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *Producer {
	return &Producer{conn: conn}
}

const (
	tracerName = "kafka-producer-adshow"
)

func (p *Producer) StartBanner(ctx context.Context, event events.BannerStart) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StartBanner")
	defer span.End()

	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicStart,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(spanCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = p.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func (p *Producer) StopBanner(ctx context.Context, event events.BannerStop) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "StopBanner")
	defer span.End()

	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicStop,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(spanCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = p.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
