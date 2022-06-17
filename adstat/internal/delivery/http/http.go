package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/crxfoz/teaserad/adstat/internal/domain"
	"github.com/crxfoz/teaserad/adstat/internal/domain/entity"
	"github.com/labstack/echo/v4"
)

type StatService interface {
	GetBannerStat(ctx context.Context, bannerID int, from time.Time) ([]*entity.BannerStat, error)
	GetBannerStatToday(ctx context.Context, bannerID int) ([]*entity.BannerStat, error)
	GetPlatformStat(ctx context.Context, platformID int, from time.Time) ([]*entity.PlatformStat, error)
	GetPlatformStatToday(ctx context.Context, platformID int) ([]*entity.PlatformStat, error)
}

type Router struct {
	statSvc StatService
	logger  domain.Logger
}

func New(statSvc StatService, logger domain.Logger) *Router {
	return &Router{statSvc: statSvc, logger: logger}
}

func (r *Router) BannerStat(c echo.Context) error {
	bb := c.Param("id")
	bannerID, err := strconv.Atoi(bb)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	from, err := time.Parse("2006-01-02", c.QueryParam("from"))
	if err != nil {
		banners, err := r.statSvc.GetBannerStatToday(c.Request().Context(), bannerID)
		if err != nil {
			r.logger.Errorw("could not get banners", "err", err, "endpoint", "BannerStat")
			return c.JSON(http.StatusInternalServerError, HTTPError{"could not get stat"})
		}

		return c.JSON(http.StatusOK, banners)

	}

	banners, err := r.statSvc.GetBannerStat(c.Request().Context(), bannerID, from)
	if err != nil {
		r.logger.Errorw("could not get banners", "err", err, "endpoint", "BannerStat")
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get stat"})
	}

	return c.JSON(http.StatusOK, banners)
}

func (r *Router) PlatformStat(c echo.Context) error {
	pp := c.Param("id")
	platformID, err := strconv.Atoi(pp)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	from, err := time.Parse("2006-01-02", c.QueryParam("from"))
	if err != nil {
		platforms, err := r.statSvc.GetPlatformStatToday(c.Request().Context(), platformID)
		if err != nil {
			r.logger.Errorw("could not get platforms", "err", err, "endpoint", "PlatformStat")
			return c.JSON(http.StatusInternalServerError, HTTPError{"could not get stat"})
		}

		return c.JSON(http.StatusOK, platforms)
	}

	platforms, err := r.statSvc.GetPlatformStat(c.Request().Context(), platformID, from)
	if err != nil {
		r.logger.Errorw("could not get platforms", "err", err, "endpoint", "PlatformStat")
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get stat"})
	}

	return c.JSON(http.StatusOK, platforms)
}
