package user

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/crxfoz/teaserad/crmad/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmad/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type Auth interface {
	Generate(user entity.User) (entity.UserContext, error)
	Verify(accessToken string) (entity.UserContext, error)
}

type Repo interface {
	AddCategory(ctx context.Context, category *entity.WebsiteCategory) error
	FindUser(ctx context.Context, username string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (int, error)
	GetBanners(ctx context.Context, userID int) ([]*entity.Banner, error)
	CreateBanner(ctx context.Context, banner *entity.Banner) (int, error)
	BannerChangeStatus(ctx context.Context, bannerID int, status bool, comment string) error
	GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error)
	GetCategories(ctx context.Context) ([]*entity.WebsiteCategory, error)
	BannerActivate(ctx context.Context, bannerID int, userID int) error
	BannerDeactivate(ctx context.Context, bannerID int, userID int) error
}

type BannerEventer interface {
	BannerCreated(ctx context.Context, msg events.BannerCreated) error
}

type BannerActor interface {
	BannerStart(ctx context.Context, msg events.BannerStart) error
	BannerStop(ctx context.Context, msg events.BannerStop) error
}

type Transactor interface {
	WithTransaction(context.Context, func(ctx context.Context) error) error
}

type User struct {
	repo          Repo
	auth          Auth
	bannerEventer BannerEventer
	bannerActor   BannerActor
	transactor    Transactor
}

func New(repo Repo, auth Auth, bannerEventer BannerEventer, bannerActor BannerActor, transactor Transactor) *User {
	return &User{repo: repo, auth: auth, bannerEventer: bannerEventer, bannerActor: bannerActor, transactor: transactor}
}

func (u *User) AddCategory(ctx context.Context, category *entity.WebsiteCategory) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddCategory")
	defer span.End()

	err := u.repo.AddCategory(spanCtx, category)
	if err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (u *User) GetCategories(ctx context.Context) ([]*entity.WebsiteCategory, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetCategories")
	defer span.End()

	list, err := u.repo.GetCategories(spanCtx)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(list) == 0 {
		return []*entity.WebsiteCategory{}, nil
	}

	return list, nil
}

func (u *User) BannerReachedLimits(ctx context.Context, item events.BannerReachedLimits) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerReachedLimits")
	defer span.End()

	// TODO: user id isn't necessary  here but I don't want to add UserID in domain
	//  event so we need extra method in repo to change status without userID valdiation
	bannerInfo, err := u.GetBanner(spanCtx, item.BannerID)
	if err != nil {
		return fmt.Errorf("could not get banner info: %w", err)
	}

	if err := u.repo.BannerDeactivate(spanCtx, item.BannerID, bannerInfo.UserID); err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (u *User) BannerStart(ctx context.Context, bannerID int, userID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerStart")
	defer span.End()

	bannerInfo, err := u.GetBanner(spanCtx, bannerID)
	if err != nil {
		return fmt.Errorf("could not get banner: %w", err)
	}

	if !bannerInfo.CanBeStarted() {
		return fmt.Errorf("banner cannot be started")
	}

	err = u.transactor.WithTransaction(spanCtx, func(txCtx context.Context) error {
		if err := u.repo.BannerActivate(txCtx, bannerID, userID); err != nil {
			return fmt.Errorf("could not activate banner: %w", err)
		}

		item := events.BannerStart{
			BannerID:    bannerID,
			UserID:      bannerInfo.UserID,
			ImgData:     bannerInfo.ImgData,
			BannerText:  bannerInfo.BannerText,
			BannerURL:   bannerInfo.BannerURL,
			LimitShows:  bannerInfo.LimitShows,
			LimitClicks: bannerInfo.LimitClicks,
			LimitBudget: bannerInfo.LimitBudget,
			Device:      bannerInfo.Device,
			CategoryID:  bannerInfo.CategoryID,
		}

		if err := u.bannerActor.BannerStart(spanCtx, item); err != nil {
			return fmt.Errorf("could not stop banner: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not execute tx: %w", err)
	}

	return nil
}

func (u *User) BannerStop(ctx context.Context, bannerID int, userID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerStop")
	defer span.End()

	bannerInfo, err := u.GetBanner(spanCtx, bannerID)
	if err != nil {
		return fmt.Errorf("could not get banner: %w", err)
	}

	if !bannerInfo.CanBeStopped() {
		return fmt.Errorf("banner already stopped")
	}

	err = u.transactor.WithTransaction(spanCtx, func(txCtx context.Context) error {
		if err := u.repo.BannerDeactivate(txCtx, bannerID, userID); err != nil {
			return fmt.Errorf("could not deactive banner: %w", err)
		}

		if err := u.bannerActor.BannerStop(spanCtx, events.BannerStop{BannerID: bannerID}); err != nil {
			return fmt.Errorf("could not stop banner: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not execute tx: %w", err)
	}

	return nil
}

func (u *User) CreateUser(ctx context.Context, user *entity.User) (int, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "CreateUser")
	defer span.End()

	newUser := &entity.User{
		Username:  user.Username,
		Password:  user.HashedPassword(),
		Validated: user.Validated,
		CreatedAt: time.Now().UTC().Unix(),
	}

	uid, err := u.repo.CreateUser(spanCtx, newUser)
	if err != nil {
		return 0, fmt.Errorf("repo failed: %w", err)
	}

	return uid, nil
}

func (u *User) Auth(ctx context.Context, username string, password string) (entity.UserContext, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "Auth")
	defer span.End()

	findedUser, err := u.repo.FindUser(spanCtx, username)
	if err != nil {
		return entity.UserContext{}, fmt.Errorf("repo failed: %w", err)
	}

	if err := findedUser.CheckPassword(password); err != nil {
		return entity.UserContext{}, fmt.Errorf("could not confirm password: %w", err)
	}

	userCtx, err := u.auth.Generate(*findedUser)
	if err != nil {
		return entity.UserContext{}, fmt.Errorf("could not generate token: %w", err)
	}

	return userCtx, nil
}

func (u *User) GetBanners(ctx context.Context, userID int) ([]*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanners")
	defer span.End()

	banners, err := u.repo.GetBanners(spanCtx, userID)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(banners) == 0 {
		return []*entity.Banner{}, nil
	}

	return banners, nil
}

const (
	tracerName = "usecase"
)

func (u *User) CreateBanner(ctx context.Context, isUserValidated bool, banner *entity.Banner) (int, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "CreateBanner")
	defer span.End()

	categories, err := u.GetCategories(spanCtx)
	if err != nil {
		return 0, fmt.Errorf("could not get categories: %w", err)
	}

	if err := banner.Validate(categories); err != nil {
		return 0, fmt.Errorf("banned not valide: %w", err)
	}

	banner.CreatedAt = time.Now().UTC().Unix()
	banner.IsValidated = isUserValidated

	imgFormat := http.DetectContentType(banner.ImgData)
	switch imgFormat {
	case "image/png":
	case "image/jpeg":
	default:
		return 0, fmt.Errorf("wrong image type: %s", imgFormat)
	}

	buff := bytes.NewBuffer(banner.ImgData)
	imgCfg, _, err := image.DecodeConfig(buff)
	if err != nil {
		return 0, fmt.Errorf("could not decome img: %w", err)
	}

	if _, ok := resolutions[resolution{w: imgCfg.Width, h: imgCfg.Height}]; !ok {
		return 0, fmt.Errorf("wrong img format: w:%d, h:%d", imgCfg.Width, imgCfg.Height)
	}

	// TODO: add transaction here? SAGA seems better cuz we dont need to make user create banner once again and
	//  with SAGA we can gurantee that eventually moderator will get banner for validation
	bannerID, err := u.repo.CreateBanner(spanCtx, banner)
	if err != nil {
		return 0, fmt.Errorf("repo failed: %w", err)
	}

	err = u.bannerEventer.BannerCreated(spanCtx, events.BannerCreated{
		BannerID:   bannerID,
		UserID:     banner.UserID,
		Validated:  isUserValidated,
		Device:     banner.Device,
		CategoryID: banner.CategoryID,
		CreatedAt:  banner.CreatedAt,
	})
	if err != nil {
		return bannerID, fmt.Errorf("could not send event that banner is created: %w", err)
	}

	return bannerID, nil
}

func (u *User) GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanner")
	defer span.End()

	banner, err := u.repo.GetBanner(spanCtx, bannerID)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	return banner, nil
}

func (u *User) BannerUpdated(ctx context.Context, updated events.BannerUpdated) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerUpdated")
	defer span.End()

	// TODO: shoud this method be in different usecases?
	// TODO: if banner was active and now got rejected we need to send msg to stop it

	if err := u.repo.BannerChangeStatus(spanCtx, updated.BannerID, updated.Valide, updated.Comment); err != nil {
		return fmt.Errorf("repo failed: could not update status: %w", err)
	}

	return nil
}
