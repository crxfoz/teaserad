package crmad

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/events"
)

const (
	topicName = "crmad.banner.updated"
)

type Banner struct {
	conn sarama.SyncProducer
}

func New(conn sarama.SyncProducer) *Banner {
	return &Banner{conn: conn}
}

func (b *Banner) BannerUpdated(msg events.BannerUpdated) error {
	out, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	_, _, err = b.conn.SendMessage(&sarama.ProducerMessage{
		Topic: topicName,
		Key:   sarama.StringEncoder("1"), // TODO: use different keys
		Value: sarama.ByteEncoder(out),
	})
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}
