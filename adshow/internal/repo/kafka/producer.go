package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adshow/internal/domain"
	"github.com/crxfoz/teaserad/adshow/internal/domain/events"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
)

const (
	topicShow = "adshow.action.show"
)

type Producer struct {
	ctx      context.Context
	cancelFn func()
	conn     sarama.SyncProducer

	queue  chan *ViewsCtx
	wg     *sync.WaitGroup
	logger domain.Logger
}

func New(ctx context.Context, conn sarama.SyncProducer, logger domain.Logger) *Producer {
	newCtx, cancelFn := context.WithCancel(ctx)

	p := &Producer{
		ctx:      newCtx,
		cancelFn: cancelFn,
		conn:     conn,
		queue:    make(chan *ViewsCtx, 10),
		wg:       &sync.WaitGroup{},
		logger:   logger,
	}

	p.wg.Add(1)
	go p.listen()

	return p
}

func (p *Producer) handleMessages(msg *ViewsCtx) {
	spanCtx, span := otel.Tracer(tracerName).Start(msg.ctx, "handleMessages")
	defer span.End()

	for _, item := range msg.views {
		if err := p.addView(spanCtx, item); err != nil {
			p.logger.Errorw("could not add view", "err", err)
		}
	}
}

func (p *Producer) listen() {
	defer p.wg.Done()

	for {
		select {
		case items, ok := <-p.queue:
			if !ok {
				return
			}

			p.handleMessages(items)
		}
	}
}

func (p *Producer) addView(ctx context.Context, view *events.View) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "addView")
	defer span.End()

	out, err := json.Marshal(view)
	if err != nil {
		return fmt.Errorf("could not marshal msg: %w", err)
	}

	pitem := &sarama.ProducerMessage{
		Topic: topicShow,
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

const (
	tracerName = "kafka-repo"
)

type ViewsCtx struct {
	views []*events.View
	ctx   context.Context
}

func (p *Producer) AddViews(ctx context.Context, views []*events.View) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddViews")
	defer span.End()

	select {
	case <-p.ctx.Done():
		return fmt.Errorf("could not add to queue")
	default:
	}

	p.queue <- &ViewsCtx{
		views: views,
		ctx:   spanCtx,
	}
	return nil
}

func (p *Producer) Stop() {
	p.cancelFn()
	close(p.queue)
	p.wg.Done()
}
