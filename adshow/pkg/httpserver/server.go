package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	httpdelivery "github.com/crxfoz/teaserad/adshow/internal/delivery/http"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

type Server struct {
	e      *echo.Echo
	router *httpdelivery.Routes
}

func New(userRouter *httpdelivery.Routes) *Server {
	e := echo.New()
	e.HideBanner = true

	return &Server{
		e:      e,
		router: userRouter,
	}
}

func (s *Server) BuildRoutes() {
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(s.e)

	userAPIV1 := s.e.Group("/api/v1")

	userAPIV1.GET("/banners", s.router.GetBanners)
	userAPIV1.POST("/platforms", s.router.AddPlatform)
	userAPIV1.GET("/view", s.router.ShowBanners)
}

func (s *Server) Start(port int) error {
	if err := s.e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}

func (s *Server) Stop() error {
	timedCtx, cancelFn := context.WithTimeout(context.Background(), time.Second*15)
	defer cancelFn()

	if err := s.e.Shutdown(timedCtx); err != nil {
		return fmt.Errorf("could not shutdown server gracefuly: %w", err)
	}

	return nil
}
