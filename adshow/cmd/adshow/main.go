package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adeliver/pkg/kafka"
	"github.com/crxfoz/teaserad/adshow/internal/delivery/http"
	kafdel "github.com/crxfoz/teaserad/adshow/internal/delivery/kafka"
	kafkaRepo "github.com/crxfoz/teaserad/adshow/internal/repo/kafka"
	"github.com/crxfoz/teaserad/adshow/internal/repo/mysql"
	tarantoolrepo "github.com/crxfoz/teaserad/adshow/internal/repo/tarantool"
	"github.com/crxfoz/teaserad/adshow/internal/services/show"
	"github.com/crxfoz/teaserad/adshow/pkg/httpserver"
	clustertnt "github.com/crxfoz/teaserad/adshow/pkg/tarantool"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/tarantool/go-tarantool"
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

	tp, err := tracer.TraceProvider("adshow", "staging", "http://jaeger:14268/api/traces")
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

	kafkaProducer, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
	}

	wrappedKafkaProducer := otelsarama.WrapSyncProducer(kafkaCfg, kafkaProducer)

	kafRep := kafkaRepo.New(context.Background(), wrappedKafkaProducer, logger.Named("kafka-producer"))

	sqlConn, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s",
			"root",
			"user123",
			"db-master",
			"3306",
			"adshow"))
	if err != nil {
		cmdLogger.Errorw("could not connect to DB", "err", err)
		return
	}

	tntConn, err := tarantool.Connect("tarantool:3301", tarantool.Opts{
		User: "admin",
		Pass: "pass",
	})
	if err != nil {
		cmdLogger.Errorw("could not connect to tarantool", "err", err)
		return
	}

	platformRepo := mysql.New(sqlConn)
	tntCluster := clustertnt.New([]*tarantool.Connection{tntConn})
	tntRepo := tarantoolrepo.New(tntCluster, logger.Named("tarantol-repo"))
	showService := show.New(tntRepo, platformRepo, kafRep)
	kafkaHandler := kafdel.New(showService)

	kafkaConsumer, err := sarama.NewConsumerGroup(kafkaBrokers, "adshow", kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not create consumer group", "err", err)
		return
	}

	kafSvc := kafka.New(logger.Named("kafka-consumer"))

	kafSvcSess, err := kafSvc.NewConsumer("adshow-consumer", kafkaConsumer, func(sess *kafka.Session) error {
		sess.AddRoute("adshow.banner.start", kafkaHandler.OnBannerStarted)
		sess.AddRoute("adshow.banner.stop", kafkaHandler.OnBannerStopped)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not build kafka session", "err", err)
		return
	}

	go func() {
		if err := kafSvcSess.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err)
		}
	}()

	httpHandler := http.New(showService, logger.Named("http-delivery"))
	httpSrv := httpserver.New(httpHandler)

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
}
