package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	kafkaDelivery "github.com/crxfoz/teaserad/adeliver/internal/delivery/kafka"
	kafkaRepo "github.com/crxfoz/teaserad/adeliver/internal/repo/kafka"
	redisRepo "github.com/crxfoz/teaserad/adeliver/internal/repo/redis"
	"github.com/crxfoz/teaserad/adeliver/internal/services/banner"
	"github.com/crxfoz/teaserad/adeliver/pkg/gateways/adshow"
	"github.com/crxfoz/teaserad/adeliver/pkg/kafka"
	redisCluster "github.com/crxfoz/teaserad/adeliver/pkg/redis"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func main() {
	z, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer z.Sync()

	logger := z.Sugar()
	cmdLogger := logger.Named("cmd")

	tp, err := tracer.TraceProvider("adeliver", "staging", "http://jaeger:14268/api/traces")
	if err != nil {
		cmdLogger.Fatalw("could not provder tracer")
		return
	}

	otel.SetTracerProvider(tp)

	// TODO: isolate
	// kafkaBrokers := []string{"0.0.0.0:9095", "localhost:9096", "localhost:9097"}
	kafkaBrokers := []string{"kafka-1:9092", "kafka-2:9092", "kafka-3:9092"}

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.Return.Successes = true

	kafkaProducerAdshow, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
		return
	}

	wrappedKafkaProducerAdshow := otelsarama.WrapSyncProducer(kafkaCfg, kafkaProducerAdshow)

	kafkaProducerRepo, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
		return
	}

	wrappedKafkaProducerRepo := otelsarama.WrapSyncProducer(kafkaCfg, kafkaProducerRepo)

	redisConns, err := getRedisConns()
	if err != nil {
		cmdLogger.Fatalw("could not connect to redis", "err", err)
		return
	}

	kafRepo := kafkaRepo.New(wrappedKafkaProducerRepo)
	adshowGateway := adshow.New(wrappedKafkaProducerAdshow)
	rCluster := redisCluster.New(redisConns)
	rRepo := redisRepo.New(rCluster)
	bannerService := banner.New(rRepo, kafRepo, adshowGateway)
	delivery := kafkaDelivery.New(bannerService)

	kafBuilder := kafka.New(logger)

	// consumer group for events
	// clicks and shows
	kafConsumerEvents, err := sarama.NewConsumerGroup(kafkaBrokers, "adeliver-events", kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not create consumer group", "err", err)
		return
	}

	kafSessEvents, err := kafBuilder.NewConsumer("adeliver-consumer-events", kafConsumerEvents, func(sess *kafka.Session) error {
		sess.AddRoute("adclick.action.click", delivery.OnActionClick)
		sess.AddRoute("adshow.action.show", delivery.OnActionView)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not setup kafka session for events", "err", err)
		return
	}

	go func() {
		if err := kafSessEvents.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err, "kind", "events")
		}
	}()

	// consumer group for banners
	// start and stop
	kafConsumerBanners, err := sarama.NewConsumerGroup(kafkaBrokers, "adeliver-banners", kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not create consumer group", "err", err)
		return
	}

	kafSessBanners, err := kafBuilder.NewConsumer("adeliver-consumer-banners", kafConsumerBanners, func(sess *kafka.Session) error {
		sess.AddRoute("adeliver.banner.start", delivery.OnBannerStarted)
		sess.AddRoute("adeliver.banner.stop", delivery.OnBannerStopped)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not setup kafka session for events", "err", err)
		return
	}

	go func() {
		if err := kafSessBanners.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err, "kind", "banners")
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		cmdLogger.Info("app stopped - signal:", s.String())
	}

	if err := kafSessBanners.Stop(); err != nil {
		cmdLogger.Errorw("could not stop consumer gracefuly", "err", err, "kind", "banners")
	}

	if err := kafSessEvents.Stop(); err != nil {
		cmdLogger.Errorw("could not stop consumer gracefuly", "err", err, "kind", "events")
	}

	if err := kafkaProducerAdshow.Close(); err != nil {
		cmdLogger.Errorw("could not stop producer gracefuly", "err", err, "kind", "adshow")
	}

	if err := kafkaProducerRepo.Close(); err != nil {
		cmdLogger.Errorw("could not stop producer gracefuly", "err", err, "kind", "repo")
	}
}
