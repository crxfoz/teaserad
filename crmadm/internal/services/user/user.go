package user

import (
	"context"
	"fmt"
	"time"

	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/crxfoz/teaserad/crmadm/internal/domain/events"
	"go.opentelemetry.io/otel"
)

type Transactor interface {
	WithTransaction(context.Context, func(ctx context.Context) error) error
}

type UserRepo interface {
	NewBanner(ctx context.Context, banner *entity.Banner) error
	GetBanners(ctx context.Context, limit int, offset int) ([]*entity.BannerResulution, error)
	GetNewBanners(ctx context.Context, limit int, offset int) ([]*entity.Banner, error)
	AddResolution(ctx context.Context, resolution *entity.Resolution) error
	AddUser(ctx context.Context, user *entity.User) error
	FindUser(ctx context.Context, username string) (*entity.User, error)
}

type Auth interface {
	Generate(user entity.User) (entity.UserContext, error)
	Verify(accessToken string) (entity.UserContext, error)
}

type BannerEventer interface {
	BannerUpdated(msg events.BannerUpdated) error
}

type User struct {
	repo          UserRepo
	bannerEventer BannerEventer
	auth          Auth
	transactor    Transactor
}

func New(auth Auth, repo UserRepo, transactor Transactor, eventer BannerEventer) *User {
	return &User{repo: repo, auth: auth, transactor: transactor, bannerEventer: eventer}
}

func (u *User) CreateUser(ctx context.Context, user *events.NewUser) error {
	if !user.CanCreateUsers() {
		return &entity.ErrForbiden{
			Role: user.MyRole,
			Msg:  "cannot create users",
		}
	}

	if !user.IsValide() {
		return &entity.ErrValidation{Msg: "wrong role"}
	}

	newUser := &entity.User{
		Username:  user.Username,
		Password:  user.HashedPassword(),
		Role:      user.Role,
		CreatedAt: time.Now().UTC().Unix(),
	}

	err := u.repo.AddUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (u *User) Auth(ctx context.Context, username string, password string) (entity.UserContext, error) {
	findedUser, err := u.repo.FindUser(ctx, username)
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

func (u *User) GetNewBanners(ctx context.Context, limit int, offset int) ([]*entity.Banner, error) {
	banners, err := u.repo.GetNewBanners(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(banners) == 0 {
		return []*entity.Banner{}, nil
	}

	return banners, nil
}

func (u *User) AddBannerResolution(ctx context.Context, resolution *entity.Resolution) error {
	err := u.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.AddResolution(txCtx, resolution); err != nil {
			return fmt.Errorf("repo failed: %w", err)
		}

		err := u.bannerEventer.BannerUpdated(events.BannerUpdated{
			BannerID: resolution.BannerID,
			Valide:   resolution.Valide,
			Comment:  resolution.Comment,
		})
		if err != nil {
			return fmt.Errorf("could not send event: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}

func (u *User) GetBanners(ctx context.Context, limit int, offset int) ([]*entity.BannerResulution, error) {
	banners, err := u.repo.GetBanners(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repo failed: %w", err)
	}

	if len(banners) == 0 {
		return []*entity.BannerResulution{}, nil
	}

	return banners, nil
}

func (u *User) NewBanner(ctx context.Context, banner events.BannerCreated) error {
	newCtx, span := otel.Tracer("usecase").Start(ctx, "NewBanner")
	defer span.End()

	if !banner.ShouldBeAdded() {
		return nil
	}

	item := &entity.Banner{
		BannerID: banner.BannerID,
		UserID:   banner.UserID,
		// TODO: Use CreatedAt from kafka
		CreatedAt: time.Now().UTC().Unix(),
	}

	if err := u.repo.NewBanner(newCtx, item); err != nil {
		return fmt.Errorf("repo failed: %w", err)
	}

	return nil
}
