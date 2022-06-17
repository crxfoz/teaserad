package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/crxfoz/teaserad/crmadm/internal/domain"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/events"
	"github.com/labstack/echo/v4"
)

type HTTPError struct {
	Error string `json:"error"`
}

type UserService interface {
	CreateUser(ctx context.Context, user *events.NewUser) error
	Auth(ctx context.Context, username string, password string) (entity.UserContext, error)
	GetNewBanners(ctx context.Context, limit int, offset int) ([]*entity.Banner, error)
	GetBanners(ctx context.Context, limit int, offset int) ([]*entity.BannerResulution, error)
	AddBannerResolution(ctx context.Context, resolution *entity.Resolution) error
}

type Routes struct {
	userSvc UserService
	logger  domain.Logger
}

func New(userSvc UserService, logger domain.Logger) *Routes {
	return &Routes{userSvc: userSvc, logger: logger}
}

func (r *Routes) NewUser(c echo.Context, userData entity.UserContext) error {
	var newUser events.NewUser
	if err := c.Bind(newUser); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong request"})
	}

	newUser.MyRole = userData.Role

	if err := r.userSvc.CreateUser(c.Request().Context(), &newUser); err != nil {
		r.logger.Errorw("could not create user",
			"endpoint", "NewUser",
			"err", err)

		var forbiddenErr entity.ErrForbiden
		if errors.Is(err, &forbiddenErr) {
			return c.JSON(http.StatusForbidden, HTTPError{Error: forbiddenErr.Error()})
		}

		var valideErr entity.ErrValidation
		if errors.Is(err, &valideErr) {
			return c.JSON(http.StatusUnprocessableEntity, HTTPError{Error: valideErr.Error()})
		}

		return c.JSON(http.StatusInternalServerError, HTTPError{"could not create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (r *Routes) Login(c echo.Context) error {
	var login LoginRequest

	if err := c.Bind(&login); err != nil {
		r.logger.Errorw("wrong request",
			"endpoint", "Login",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	userCtx, err := r.userSvc.Auth(c.Request().Context(), login.Username, login.Password)
	if err != nil {
		r.logger.Errorw("could not login",
			"endpoint", "Login",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not login"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": userCtx.Token,
	})
}

func (r *Routes) limitAndOffset(c echo.Context, defaultLimit int, defaultOffset int) (limit int, offset int) {
	l := c.QueryParam("limit")
	limit = defaultLimit

	if lInt, err := strconv.Atoi(l); err == nil {
		limit = lInt
	}

	o := c.QueryParam("offset")
	offset = defaultOffset

	if oInt, err := strconv.Atoi(o); err == nil {
		offset = oInt
	}

	return
}

func (r *Routes) GetNewBanners(c echo.Context, userData entity.UserContext) error {
	limit, offset := r.limitAndOffset(c, 10, 0)

	banners, err := r.userSvc.GetNewBanners(c.Request().Context(), limit, offset)
	if err != nil {
		r.logger.Errorw("could not get banners",
			"endpoint", "GetNewBanners",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get banners"})
	}

	return c.JSON(http.StatusOK, banners)
}

func (r *Routes) NewResolution(c echo.Context, userData entity.UserContext) error {
	var resolution ResolutinRequest

	if err := c.Bind(&resolution); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	res := &entity.Resolution{
		BannerID: resolution.BannerID,
		Valide:   resolution.Valide,
		Comment:  resolution.Comment,
	}
	if err := r.userSvc.AddBannerResolution(c.Request().Context(), res); err != nil {
		r.logger.Errorw("could not add resolution",
			"endpoint", "NewResolution",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not add resolution"})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"status": "ok",
	})
}

func (r *Routes) GetBanners(c echo.Context, userData entity.UserContext) error {
	limit, offset := r.limitAndOffset(c, 10, 0)

	banners, err := r.userSvc.GetBanners(c.Request().Context(), limit, offset)
	if err != nil {
		r.logger.Errorw("could not get banners",
			"endpoint", "GetBanners",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get banners"})
	}

	return c.JSON(http.StatusOK, banners)
}
