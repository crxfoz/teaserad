package http

import (
	"context"
	"fmt"

	"github.com/crxfoz/teaserad/crmad/internal/domain"
	"github.com/crxfoz/teaserad/crmad/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmad/pkg/auth/middleware"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

// type authMiddleware *middleware.AuthMiddleware[entity.User, entity.UserContext]

type UserService interface {
	AddCategory(ctx context.Context, category *entity.WebsiteCategory) error
	CreateUser(ctx context.Context, user *entity.User) (int, error)
	Auth(ctx context.Context, username string, password string) (entity.UserContext, error)
	GetBanners(ctx context.Context, userID int) ([]*entity.Banner, error)
	CreateBanner(ctx context.Context, isUserValidated bool, banner *entity.Banner) (int, error)
	GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error)
	BannerStart(ctx context.Context, bannerID int, userID int) error
	BannerStop(ctx context.Context, bannerID int, userID int) error
	GetCategories(ctx context.Context) ([]*entity.WebsiteCategory, error)
}

type Server struct {
	ctx            context.Context
	authMiddleware *middleware.AuthMiddleware[entity.User, entity.UserContext]
	userSvc        UserService
	e              *echo.Echo
	logger         domain.Logger
}

func New(ctx context.Context, authMiddleware *middleware.AuthMiddleware[entity.User, entity.UserContext], userSvc UserService, logger domain.Logger) *Server {
	e := echo.New()
	e.HideBanner = true

	return &Server{
		ctx:            ctx,
		authMiddleware: authMiddleware,
		userSvc:        userSvc,
		e:              e,
		logger:         logger,
	}
}

func (s *Server) Run(port int) error {
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(s.e)

	s.e.GET("/static/banner/:id", s.ShowBannerImg)

	apiV1 := s.e.Group("/api/v1")

	apiV1.POST("/register", s.UserRegister)
	apiV1.POST("/login", s.UserLogin)
	apiV1.GET("/banners", s.authMiddleware.Do(s.GetBanners))
	apiV1.POST("/banners", s.authMiddleware.Do(s.AddBanner))
	apiV1.POST("/banners/start", s.authMiddleware.Do(s.BannerStart))
	apiV1.POST("/banners/stop", s.authMiddleware.Do(s.BannerStop))
	apiV1.POST("/categories", s.authMiddleware.Do(s.AddCategory))
	apiV1.GET("/categories", s.authMiddleware.Do(s.GetCategories))

	return s.e.Start(fmt.Sprintf(":%d", port))
}
