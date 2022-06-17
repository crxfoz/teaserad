package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/crxfoz/teaserad/adshow/internal/domain"
	"github.com/crxfoz/teaserad/adshow/internal/domain/entity"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

type ShowService interface {
	AddPlatform(ctx context.Context, platform *entity.Platform) error
	ShowBanners(ctx context.Context, limit int, hitCtx *entity.HitContext) ([]*entity.Banner, error)
	GetBannersForPlatform(ctx context.Context, platformID int, deviceType string, limit int) ([]*entity.Banner, error)
}

type Routes struct {
	showService ShowService
	logger      domain.Logger
}

func New(showService ShowService, logger domain.Logger) *Routes {
	return &Routes{showService: showService, logger: logger}
}

type addPlatform struct {
	PlatformID int `json:"platform_id"`
	CategoryID int `json:"category_id"`
}

const (
	tracerName = "http-delivery"
)

func (r *Routes) GetBanners(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "GetBanners")
	defer span.End()

	pid := c.QueryParam("pid")
	platformID, err := strconv.Atoi(pid)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong id"})
	}

	l := c.QueryParam("limit")
	limit, err := strconv.Atoi(l)
	if err != nil {
		limit = 10
	}

	d := c.QueryParam("device")

	banners, err := r.showService.GetBannersForPlatform(spanCtx, platformID, d, limit)
	if err != nil {
		r.logger.Errorw("could not get banners", "err", err, "endpoint", "GetBanners")
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get banners"})
	}

	return c.JSON(http.StatusOK, banners)
}

func (r *Routes) ShowBanners(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "ShowBanners")
	defer span.End()

	pid := c.QueryParam("pid")
	platformID, err := strconv.Atoi(pid)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong id"})
	}

	l := c.QueryParam("limit")
	limit, err := strconv.Atoi(l)
	if err != nil {
		limit = 1
	}

	banners, err := r.showService.ShowBanners(spanCtx, limit, &entity.HitContext{
		UserAgent:  c.Request().UserAgent(),
		PlatformID: platformID,
	})
	if err != nil {
		r.logger.Errorw("could not get banners", "err", err, "endpoint", "ShowBanners")
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get banners"})
	}

	return c.JSON(http.StatusOK, banners)
}

func (r *Routes) AddPlatform(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "AddPlatform")
	defer span.End()

	var platform addPlatform
	if err := c.Bind(&platform); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	err := r.showService.AddPlatform(spanCtx, &entity.Platform{
		PlatformID: platform.PlatformID,
		CategoryID: platform.CategoryID,
	})

	if err != nil {
		r.logger.Errorw("could not add platform", "err", err, "endpoint", "AddPlatform")

		return c.JSON(http.StatusInternalServerError, HTTPError{"could not add platform"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
