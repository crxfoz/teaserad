package http

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/crxfoz/teaserad/crmad/internal/domain/entity"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

func (s *Server) ShowBannerImg(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "ShowBannerImg")
	defer span.End()

	// TODO: other user may get banner img of other users. Can fix it using some hashing instead of just an ID
	v, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, HTTPError{"wrong img id"})
	}

	banner, err := s.userSvc.GetBanner(spanCtx, v)
	if err != nil {
		s.logger.Errorw("could not get banner", "endpoint", "ShowBannerImg", "err", err)
		return c.NoContent(http.StatusNotFound)
	}

	contentType := http.DetectContentType(banner.ImgData)

	return c.Blob(http.StatusOK, contentType, banner.ImgData)
}

func (s *Server) UserRegister(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "UserRegister")
	defer span.End()

	var register Register

	if err := c.Bind(&register); err != nil {
		s.logger.Errorw("wrong request", "endpoint", "UserRegister")
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	uid, err := s.userSvc.CreateUser(spanCtx, &entity.User{
		Username:  register.Username,
		Password:  register.Password,
		Validated: false,
	})
	if err != nil {
		s.logger.Errorw("could not register", "endpoint", "UserRegister", "err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not register"})
	}

	return c.JSON(http.StatusCreated, map[string]int{
		"id": uid,
	})
}

const (
	tracerName = "http-delivery"
)

func (s *Server) UserLogin(c echo.Context) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "UserLogin")
	defer span.End()

	var login Login

	if err := c.Bind(&login); err != nil {
		s.logger.Errorw("wrong request",
			"endpoint", "UserLogin",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	userCtx, err := s.userSvc.Auth(spanCtx, login.Username, login.Password)
	if err != nil {
		s.logger.Errorw("could not login",
			"endpoint", "UserLogin",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not login"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": userCtx.Token,
	})
}

func (s *Server) GetBanners(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "GetBanners")
	defer span.End()

	banners, err := s.userSvc.GetBanners(spanCtx, userCtx.ID)
	if err != nil {
		s.logger.Errorw("could not get banners",
			"endpoint", "GetBanners",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get banners"})
	}

	return c.JSON(http.StatusOK, banners)
}

func (s *Server) AddBanner(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "AddBanner")
	defer span.End()

	var bannerData NewBanner

	if err := c.Bind(&bannerData); err != nil {
		s.logger.Errorw("wrong request",
			"endpoint", "AddBanner",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	if bannerData.ImgData == "" {
		return c.JSON(http.StatusUnsupportedMediaType, HTTPError{"wrong img"})
	}

	imgData, err := base64.StdEncoding.DecodeString(bannerData.ImgData)
	if err != nil {
		s.logger.Errorw("could not decode img",
			"endpoint", "AddBanner",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not decode image"})
	}

	bannerID, err := s.userSvc.CreateBanner(spanCtx, userCtx.Validated, &entity.Banner{
		UserID:      userCtx.ID,
		ImgData:     imgData,
		BannerText:  bannerData.BannerText,
		BannerURL:   bannerData.BannerURL,
		IsActive:    false,
		LimitShows:  bannerData.LimitShows,
		LimitClicks: bannerData.LimitClicks,
		LimitBudget: bannerData.LimitBudget,
		CategoryID:  bannerData.CategoryID,
		Device:      bannerData.Device,
	})

	if err != nil {
		s.logger.Errorw("could not create banner",
			"endpoint", "AddBanner",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not create banner"})
	}

	return c.JSON(http.StatusOK, map[string]int{
		"banner_id": bannerID,
	})
}

type newCategory struct {
	Name string `json:"name"`
}

func (s *Server) GetCategories(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "GetCategories")
	defer span.End()

	categories, err := s.userSvc.GetCategories(spanCtx)
	if err != nil {
		s.logger.Errorw("could not get categories",
			"endpoint", "GetCategories",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not get categories"})
	}

	return c.JSON(http.StatusOK, categories)
}

func (s *Server) AddCategory(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "AddCategory")
	defer span.End()

	var categoryData newCategory

	if err := c.Bind(&categoryData); err != nil {
		s.logger.Errorw("wrong request",
			"endpoint", "AddCategory",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong data"})
	}

	err := s.userSvc.AddCategory(spanCtx, &entity.WebsiteCategory{Name: categoryData.Name})

	if err != nil {
		s.logger.Errorw("could not create banner",
			"endpoint", "AddCategory",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not create banner"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

type BannerStart struct {
	BannerID int `json:"banner_id"`
}

type BannerStop struct {
	BannerID int `json:"banner_id"`
}

func (s *Server) BannerStart(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "BannerStart")
	defer span.End()

	var start BannerStart

	if err := c.Bind(&start); err != nil {
		s.logger.Errorw("wrong data",
			"endpoint", "BannerStart",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong entity"})
	}

	err := s.userSvc.BannerStart(spanCtx, start.BannerID, userCtx.ID)
	if err != nil {
		s.logger.Errorw("could not start banner",
			"endpoint", "BannerStart",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not start banner"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
func (s *Server) BannerStop(c echo.Context, userCtx entity.UserContext) error {
	spanCtx, span := otel.Tracer(tracerName).Start(c.Request().Context(), "BannerStop")
	defer span.End()

	var stop BannerStop

	if err := c.Bind(&stop); err != nil {
		s.logger.Errorw("wrong data",
			"endpoint", "BannerStop",
			"err", err)
		return c.JSON(http.StatusUnprocessableEntity, HTTPError{"wrong entity"})
	}

	err := s.userSvc.BannerStop(spanCtx, stop.BannerID, userCtx.ID)
	if err != nil {
		s.logger.Errorw("could not stop banner",
			"endpoint", "BannerStop",
			"err", err)
		return c.JSON(http.StatusInternalServerError, HTTPError{"could not stop banner"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
