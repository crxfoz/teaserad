package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/crxfoz/teaserad/crmad/pkg/auth/middleware"
	httpdelivery "github.com/crxfoz/teaserad/crmadm/internal/delivery/http"
	"github.com/crxfoz/teaserad/crmadm/internal/domain"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

type Server struct {
	e              *echo.Echo
	logger         domain.Logger
	authMiddleware *middleware.AuthMiddleware[entity.User, entity.UserContext]
	router         *httpdelivery.Routes
}

func New(userRouter *httpdelivery.Routes, authMiddleware *middleware.AuthMiddleware[entity.User, entity.UserContext]) *Server {
	e := echo.New()
	e.HideBanner = true

	return &Server{
		e:              e,
		authMiddleware: authMiddleware,
		router:         userRouter,
	}
}

func (s *Server) BuildRoutes() {
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(s.e)

	userAPIV1 := s.e.Group("/api/v1")

	userAPIV1.GET("/banners/new", s.authMiddleware.Do(s.router.GetNewBanners))
	userAPIV1.GET("/banners", s.authMiddleware.Do(s.router.GetBanners))
	userAPIV1.POST("/resolution", s.authMiddleware.Do(s.router.NewResolution))
	userAPIV1.POST("/login", s.router.Login)
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
