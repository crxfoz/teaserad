package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/crxfoz/teaserad/adeliver/pkg/kafka"
	"github.com/crxfoz/teaserad/crmad/internal/delivery/http"
	kafkaController "github.com/crxfoz/teaserad/crmad/internal/delivery/kafka"
	"github.com/crxfoz/teaserad/crmad/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmad/internal/repo/mysql"
	"github.com/crxfoz/teaserad/crmad/internal/services/jwt"
	"github.com/crxfoz/teaserad/crmad/internal/services/user"
	"github.com/crxfoz/teaserad/crmad/pkg/auth/middleware"
	"github.com/crxfoz/teaserad/crmad/pkg/gateways/adeliver"
	"github.com/crxfoz/teaserad/crmad/pkg/gateways/crmadm"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

	tp, err := tracer.TraceProvider("crmad", "staging", "http://jaeger:14268/api/traces")
	if err != nil {
		cmdLogger.Fatalw("could not provder tracer")
		return
	}

	otel.SetTracerProvider(tp)

	// TODO: isolate
	// kafkaBrokers := []string{"kafka-1:9094", "kafka-2:9094", "kafka-3:9094"}
	kafkaBrokers := []string{"kafka-1:9092", "kafka-2:9092", "kafka-3:9092"}
	// kafkaBrokers := []string{"0.0.0.0:9095", "0.0.0.0:9096", "0.0.0.0:9097"}
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.Return.Successes = true
	kafkaProducer, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
	}

	wrappedKafkaProducer := otelsarama.WrapSyncProducer(kafkaCfg, kafkaProducer)
	crmAdmGateway := crmadm.New(wrappedKafkaProducer)

	kafkaProducerAdeliver, err := sarama.NewSyncProducer(kafkaBrokers, kafkaCfg)
	if err != nil {
		cmdLogger.Fatalw("could not connect to kafka",
			"err", err)
	}
	wrappedkafkaProducerAdeliver := otelsarama.WrapSyncProducer(kafkaCfg, kafkaProducerAdeliver)

	adeliverGateway := adeliver.New(wrappedkafkaProducerAdeliver)

	sqlConn, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s",
			"root",
			"user123",
			"db-master",
			"3306",
			"crmad"))
	if err != nil {
		cmdLogger.Errorw("could not connect to DB", "err", err)
		return
	}

	authManager := jwt.NewJWTManager("123", time.Hour*24*7) // TODO: add token to envs
	userRepo := mysql.New(sqlConn)
	userSvc := user.New(userRepo, authManager, crmAdmGateway, adeliverGateway, userRepo)
	authMiddleware := middleware.New[entity.User, entity.UserContext](authManager)
	srv := http.New(context.Background(), authMiddleware, userSvc, logger.Named("crmad-delivery-http"))

	go func() {
		if err := srv.Run(8080); err != nil {
			cmdLogger.Errorw("server stopped", "err", err)
		}
	}()

	kafBuilder := kafka.New(logger)
	kfController := kafkaController.New(userSvc)

	// consumer to listen events from crmadm service
	kafConsumerCrmad, err := sarama.NewConsumerGroup(kafkaBrokers, "crmad", kafkaCfg)
	if err != nil {
		cmdLogger.Errorw("could not create consumer group", "err", err)
		return
	}

	kafAdminConsumer, err := kafBuilder.NewConsumer("crmad-consumer", kafConsumerCrmad, func(sess *kafka.Session) error {
		sess.AddRoute("crmad.banner.updated", kfController.OnBannerUpdated)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not start consumer", "err", err, "kind", "crmadm")
		return
	}

	go func() {
		if err := kafAdminConsumer.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err)
		}
	}()

	// consumer to listen events from crmadm service
	kafConsumerAdeliver, err := sarama.NewConsumerGroup(kafkaBrokers, "crmad-adeliver", kafkaCfg)
	if err != nil {
		cmdLogger.Errorw("could not create consumer group", "err", err)
		return
	}

	kafAdeliverConsumer, err := kafBuilder.NewConsumer("crmad-consumer", kafConsumerAdeliver, func(sess *kafka.Session) error {
		sess.AddRoute("adeliver.banner.limits", kfController.OnBannerReachedLimits)
		return nil
	})
	if err != nil {
		cmdLogger.Fatalw("could not start consumer", "err", err, "kind", "adeliver")
		return
	}

	go func() {
		if err := kafAdeliverConsumer.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		cmdLogger.Info("app stopped - signal:", s.String())
	}

	if err := kafAdeliverConsumer.Stop(); err != nil {
		cmdLogger.Errorw("could not stop consumer gracefuly", "err", err, "kind", "adeliver")
	}

	if err := kafAdminConsumer.Stop(); err != nil {
		cmdLogger.Errorw("could not stop consumer gracefuly", "err", err, "kind", "crmadm")
	}

	if err := kafkaProducerAdeliver.Close(); err != nil {
		cmdLogger.Errorw("could not stop producer gracefuly", "err", err, "kind", "adeliver")
	}

	if err := kafkaProducer.Close(); err != nil {
		cmdLogger.Errorw("could not stop producer gracefuly", "err", err, "kind", "crmadm")
	}

	if err := tp.Shutdown(context.Background()); err != nil {
		cmdLogger.Errorw("could not stop tracing gracefuly", "err", err)
	}
}
