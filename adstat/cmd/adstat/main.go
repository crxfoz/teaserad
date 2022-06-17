package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/crxfoz/teaserad/adstat/internal/delivery/http"
	"github.com/crxfoz/teaserad/adstat/internal/repo/clickhouse"
	"github.com/crxfoz/teaserad/adstat/internal/services/stat"
	"github.com/crxfoz/teaserad/adstat/pkg/httpserver"
	"github.com/crxfoz/teaserad/crmad/pkg/tracer"
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

	tp, err := tracer.TraceProvider("adstat", "staging", "http://jaeger:14268/api/traces")
	if err != nil {
		cmdLogger.Fatalw("could not provder tracer")
		return
	}

	otel.SetTracerProvider(tp)

	sqlConn, err := sqlx.Connect("clickhouse",
		fmt.Sprintf("clickhouse://%s:%s/stat?dial_timeout=200ms&max_execution_time=60",
			"clickhouse-1",
			"9000"))
	if err != nil {
		cmdLogger.Errorw("could not connect to DB", "err", err)
		return
	}

	repo := clickhouse.New(sqlConn)
	statService := stat.New(repo, logger.Named("adstat-service"))
	httpHandler := http.New(statService, logger.Named("adstat-delivery-http"))

	httpSrv := httpserver.New(httpHandler)

	go func() {
		httpSrv.BuildRoutes()

		if err := httpSrv.Start(8080); err != nil {
			cmdLogger.Errorw("server stopped", "err", err)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		cmdLogger.Info("app stopped - signal:", s.String())
	}

	if err := httpSrv.Stop(); err != nil {
		cmdLogger.Errorw("could not stop http-server gracefuly", "err", err)
	}
}
