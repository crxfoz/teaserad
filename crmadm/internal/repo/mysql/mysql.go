package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

type executor interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type txKey struct{}

func injectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}

	return nil
}

func (r *UserRepo) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	err = fn(injectTx(ctx, tx))
	if err != nil {
		if errLoc := tx.Rollback(); errLoc != nil {
			// TODO: custom error type
			return fmt.Errorf("could not rollback tx: %w, original reason: %s", errLoc, err.Error())
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit tx: %w", err)
	}

	return nil
}

func (r *UserRepo) executor(ctx context.Context) executor {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}

	return r.db
}

func (r *UserRepo) GetBanners(ctx context.Context, limit int, offset int) ([]*entity.BannerResulution, error) {
	conn := r.executor(ctx)
	var banners []*entity.BannerResulution

	err := conn.GetContext(ctx, &banners, `SELECT b.banner_id, b.user_id, b.created_at, r.valide, r.comment, r.created_at as updated_at FROM banner b JOIN resolution r ON b.banner_id=r.banner_id ORDER BY b.created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}

	return banners, nil
}

func (r *UserRepo) AddUser(ctx context.Context, user *entity.User) error {
	conn := r.executor(ctx)

	_, err := conn.ExecContext(ctx, "INSERT INTO user (username, password, role, created_at) VALUES (?,?,?,?)",
		user.Username,
		user.Password,
		user.Role,
		user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not insert resolution: %w", err)
	}

	return nil
}

func (r *UserRepo) AddResolution(ctx context.Context, resolution *entity.Resolution) error {
	conn := r.executor(ctx)

	_, err := conn.ExecContext(ctx, "INSERT INTO resolution (banner_id, valide, comment, created_at) VALUES (?,?,?,?)",
		resolution.BannerID,
		resolution.Valide,
		resolution.Comment,
		resolution.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not insert resolution: %w", err)
	}

	return nil
}

func (r *UserRepo) GetNewBanners(ctx context.Context, limit int, offset int) ([]*entity.Banner, error) {
	conn := r.executor(ctx)

	var banners []*entity.Banner
	err := conn.SelectContext(ctx, &banners, `SELECT b.banner_id,b.user_id,b.created_at 
		FROM banner b WHERE NOT EXISTS(
			SELECT r.banner_id FROM resolution r WHERE b.banner_id=r.banner_id)
		ORDER BY b.created_at ASC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("could not select: %w", err)
	}

	return banners, nil
}

func (r *UserRepo) FindUser(ctx context.Context, username string) (*entity.User, error) {
	conn := r.executor(ctx)

	var user entity.User
	if err := conn.GetContext(ctx, &user, `SELECT id, username, password, role, created_at FROM user WHERE username=?`, username); err != nil {
		return nil, fmt.Errorf("could not get user from mysql: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) NewBanner(ctx context.Context, banner *entity.Banner) error {
	newCtx, span := otel.Tracer("db").Start(ctx, "NewBanner")
	defer span.End()

	conn := r.executor(newCtx)

	_, err := conn.ExecContext(newCtx, "INSERT INTO banner (banner_id, user_id, created_at) VALUES (?,?,?)", banner.BannerID, banner.UserID, banner.CreatedAt)
	if err != nil {
		return fmt.Errorf("could not insert new banner: %w", err)
	}

	return nil
}
