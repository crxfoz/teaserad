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
	"github.com/crxfoz/teaserad/crmad/pkg/auth/middleware"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
	httpController "github.com/crxfoz/teaserad/crmadm/internal/delivery/http"
	kafkaController "github.com/crxfoz/teaserad/crmadm/internal/delivery/kafka"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmadm/internal/repo/mysql"
	"github.com/crxfoz/teaserad/crmadm/internal/services/jwt"
	"github.com/crxfoz/teaserad/crmadm/internal/services/user"
	"github.com/crxfoz/teaserad/crmadm/pkg/gateways/crmad"
	"github.com/crxfoz/teaserad/crmadm/pkg/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

	tp, err := tracer.TraceProvider("crmadmin", "staging", "http://jaeger:14268/api/traces")
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

	crmadGateway := crmad.New(kafkaProducer)

	sqlConn, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s",
			"root",
			"user123",
			"db-master",
			"3306",
			"crmadm"))
	if err != nil {
		cmdLogger.Errorw("could not connect to DB", "err", err)
		return
	}

	authManager := jwt.NewJWTManager("123", time.Hour*24*7) // TODO: add token to envs
	userRepo := mysql.NewRepo(sqlConn)
	userSvc := user.New(authManager, userRepo, userRepo, crmadGateway)
	authMiddleware := middleware.New[entity.User, entity.UserContext](authManager)

	userHTTPRouter := httpController.New(userSvc, logger.Named("crmadm-delivery-http"))
	httpSrv := server.New(userHTTPRouter, authMiddleware)

	kafkaConsumer, err := sarama.NewConsumerGroup(kafkaBrokers, "crmadm", kafkaCfg)
	if err != nil {
		cmdLogger.Errorw("could not create consumer group", "err", err)
		return
	}

	kfController := kafkaController.New(userSvc)
	kafBuilder := kafka.New(logger)

	kafkaSess, err := kafBuilder.NewConsumer("crmadm-delivery-kafka", kafkaConsumer, func(session *kafka.Session) error {
		session.AddRoute("crmadm.banner.created", kfController.OnNewBanner)
		return nil
	})

	go func() {
		if err := kafkaSess.Start(); err != nil {
			cmdLogger.Errorw("kafka listener stopped", "err", err)
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

	if err := tp.Shutdown(context.Background()); err != nil {
		cmdLogger.Errorw("could not stop tracing gracefuly", "err", err)
	}
}
