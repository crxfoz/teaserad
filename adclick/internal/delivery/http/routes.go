package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/crxfoz/teaserad/adclick/internal/domain"
	"github.com/crxfoz/teaserad/adclick/internal/domain/entity"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

type ClickService interface {
	NewClick(ctx context.Context, bannerID int, platformID int, viewID string) (*entity.BannerURL, error)
}

type Router struct {
	clickSvc ClickService
	logger   domain.Logger
}

func New(clickSvc ClickService, logger domain.Logger) *Router {
	return &Router{clickSvc: clickSvc, logger: logger}
}

const (
	tracerName = "http-delivery"
)

func (r *Router) RegisterClick(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "RegisterClick")
	defer span.End()

	pid := c.QueryParam("pid")
	platformID, err := strconv.Atoi(pid)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"invalide platform id"})
	}

	bid := c.QueryParam("bid")
	bannerID, err := strconv.Atoi(bid)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"invalide banner id"})
	}

	// TODO: add viewID
	info, err := r.clickSvc.NewClick(spanCtx, bannerID, platformID, "")
	if err != nil {
		r.logger.Errorw("could not register click", "err", err, "endpoint", "RegisterClick")
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not register new click"})
	}

	return c.JSON(http.StatusOK, info)
}

func (r *Router) HTMLClick(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "HTMLClick")
	defer span.End()

	pid := c.QueryParam("pid")
	platformID, err := strconv.Atoi(pid)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	bid := c.QueryParam("bid")
	bannerID, err := strconv.Atoi(bid)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	// TODO: add viewID
	info, err := r.clickSvc.NewClick(spanCtx, bannerID, platformID, "")
	if err != nil {
		r.logger.Errorw("could not register click", "err", err, "endpoint", "HTMLClick")
		return c.NoContent(http.StatusNotFound)
	}

	return c.Redirect(http.StatusMovedPermanently, info.BannerURL)
}
