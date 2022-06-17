package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adclick/internal/delivery/http"
	kafkadel "github.com/crxfoz/teaserad/adclick/internal/delivery/kafka"
	kafrepo "github.com/crxfoz/teaserad/adclick/internal/repo/kafka"
	redisRepo "github.com/crxfoz/teaserad/adclick/internal/repo/redis"
	"github.com/crxfoz/teaserad/adclick/internal/services/click"
	"github.com/crxfoz/teaserad/adclick/pkg/httpserver"
	"github.com/crxfoz/teaserad/adeliver/pkg/kafka"
	redisCluster "github.com/crxfoz/teaserad/adeliver/pkg/redis"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
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

	tp, err := tracer.TraceProvider("adclick", "staging", "http://jaeger:14268/api/traces")
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

	kafkaProducerRepo, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
		return
	}

	redisConns, err := getRedisConns()
	if err != nil {
		cmdLogger.Fatalw("could not connect to redis", "err", err)
		return
	}

	kafConsumerNewBanners, err := sarama.NewConsumerGroup(kafkaBrokers, "adclick-banners", kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not create consumer group", "err", err)
		return
	}

	kafkaRepo := kafrepo.New(kafkaProducerRepo)
	rCluster := redisCluster.New(redisConns)
	rRepo := redisRepo.New(rCluster)
	clickService := click.New(rRepo, kafkaRepo, logger.Named("service-click"))
	kafBuilder := kafka.New(logger)
	kafkaHandler := kafkadel.New(clickService)
	httpHandler := http.New(clickService, logger.Named("delivery-http"))
	httpSrv := httpserver.New(httpHandler)

	kafSessBanners, err := kafBuilder.NewConsumer("adclick-consumer-banners", kafConsumerNewBanners, func(sess *kafka.Session) error {
		sess.AddRoute("adshow.banner.start", kafkaHandler.OnNewBanner)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not setup kafka session for events", "err", err)
		return
	}

	go func() {
		if err := kafSessBanners.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err, "kind", "events")
		}
	}()

	go func() {
		httpSrv.BuildRoutes()

		if err := httpSrv.Start(8080); err != nil {
			cmdLogger.Errorw("http server stopped", "err", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		cmdLogger.Info("app stopped - signal:", s.String())
	}

	if err := httpSrv.Stop(); err != nil {
		cmdLogger.Errorw("could not stop http-server", "err", err)
	}

	if err := kafSessBanners.Stop(); err != nil {
		cmdLogger.Errorw("could not stop consumer gracefuly", "err", err, "kind", "banners")
	}

	if err := kafkaProducerRepo.Close(); err != nil {
		cmdLogger.Errorw("could not stop producer gracefuly", "err", err, "kind", "repo")
	}
}
