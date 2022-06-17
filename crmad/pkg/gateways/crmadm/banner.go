package crmadm

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
	topicName = "crmadm.banner.created"
)

type Banner struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *Banner {
	return &Banner{conn: conn}
}

const (
	tracerName = "kafka-producer-crmadm"
)

func (b *Banner) BannerCreated(ctx context.Context, msg events.BannerCreated) error {
	newCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerCreated")
	defer span.End()

	out, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicName,
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
