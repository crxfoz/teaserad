package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adeliver/internal/domain"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type IncomeHandler struct {
	logger domain.Logger
	routes map[string]RouteFn
}

func (i *IncomeHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (i *IncomeHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (i *IncomeHandler) handleMsg(item *sarama.ConsumerMessage, fn RouteFn) {
	prevCtx := otel.GetTextMapPropagator().Extract(context.Background(), otelsarama.NewConsumerMessageCarrier(item))
	ctx, span := otel.Tracer("consumer").Start(prevCtx, "OnMessage")
	defer span.End()

	if err := fn(ctx, item); err != nil {
		i.logger.Errorw("could not process msg", "err", err)
		return
	}
}

func (i *IncomeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for item := range claim.Messages() {
		if routeFn, ok := i.routes[item.Topic]; ok {
			i.handleMsg(item, routeFn)
		} else {
			i.logger.Warnw("got unexpected msg", "topic", item.Topic)
		}

		// TODO: ack?
		// session.MarkMessage(item, "commited")
	}

	return nil
}

type RouteFn func(context.Context, *sarama.ConsumerMessage) error

type Kafka struct {
	logger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger) *Kafka {
	return &Kafka{logger: logger}
}

type Session struct {
	routes  map[string]RouteFn
	handler sarama.ConsumerGroupHandler
	conn    sarama.ConsumerGroup
}

func newSession(conn sarama.ConsumerGroup) *Session {
	return &Session{
		routes: make(map[string]RouteFn),
		conn:   conn,
	}
}

func (s *Session) topics() []string {
	out := make([]string, 0, len(s.routes))

	for topic := range s.routes {
		out = append(out, topic)
	}

	return out
}
func (s *Session) Start() error {
	if err := s.conn.Consume(context.Background(), s.topics(), s.handler); err != nil {
		return fmt.Errorf("could not setup consumer: %w", err)
	}

	return nil
}

func (s *Session) Stop() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("could not close consumer: %w", err)
	}

	return nil
}

func (s *Session) AddRoute(topic string, route RouteFn) {
	s.routes[topic] = route
}

func (k *Kafka) NewConsumer(withName string, conn sarama.ConsumerGroup, sessBuilder func(session *Session) error) (*Session, error) {
	sess := newSession(conn)
	if err := sessBuilder(sess); err != nil {
		return nil, fmt.Errorf("could not build session: %w", err)
	}

	handler := &IncomeHandler{
		logger: k.logger.Named(withName),
		routes: sess.routes,
	}

	wrappedHandler := otelsarama.WrapConsumerGroupHandler(handler)

	sess.handler = wrappedHandler

	return sess, nil
}
