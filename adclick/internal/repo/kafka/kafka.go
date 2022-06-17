package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adclick/internal/domain/events"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
)

const (
	topicNewClick = "adclick.action.click"
	tracerName    = "kafka-producer"
)

type Kafka struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *Kafka {
	return &Kafka{conn: conn}
}

func (k *Kafka) SendClick(ctx context.Context, event *events.Click) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "SendClick")
	defer span.End()

	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicNewClick,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	}

	otel.GetTextMapPropagator().Inject(spanCtx, otelsarama.NewProducerMessageCarrier(pitem))

	_, _, err = k.conn.SendMessage(pitem)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
