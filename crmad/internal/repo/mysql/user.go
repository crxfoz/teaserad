package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/crxfoz/teaserad/crmad/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

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

func (ur *UserRepo) executor(ctx context.Context) executor {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}

	return ur.db
}

type UserRepo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (ur *UserRepo) FindUser(ctx context.Context, username string) (*entity.User, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "FindUser")
	defer span.End()

	var user entity.User

	if err := ur.db.GetContext(spanCtx, &user, `SELECT id, username, password, validated, created_at FROM users WHERE username=?`, username); err != nil {
		return nil, fmt.Errorf("could not get user from mysql: %w", err)
	}

	return &user, nil
}

func (ur *UserRepo) CreateUser(ctx context.Context, user *entity.User) (int, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "CreateUser")
	defer span.End()

	tx, err := ur.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("could not start tx: %w", err)
	}

	_, err = tx.ExecContext(spanCtx, `INSERT INTO users (username, password, validated, created_at) VALUES (?,?,?,?)`,
		user.Username,
		user.Password,
		user.Validated,
		user.CreatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("could not insert user to mysql: %w", err)
	}

	var userID int
	if err := tx.GetContext(spanCtx, &userID, "SELECT LAST_INSERT_ID()"); err != nil {
		if err := tx.Rollback(); err != nil {
			return 0, fmt.Errorf("could not rollback tx: %w", err)
		}
		return 0, fmt.Errorf("could not get id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit: %w", err)
	}

	return userID, nil
}

func (ur *UserRepo) GetBanners(ctx context.Context, userID int) ([]*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanners")
	defer span.End()

	var banners []*entity.Banner

	err := ur.db.SelectContext(spanCtx, &banners,
		`SELECT id, img_data, banner_text, banner_url, is_active, limit_shows,
       		limit_clicks, limit_budget, user_id, created_at, is_validated, comment, device, category_id
		FROM banners WHERE user_id=?`, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get banners: %w", err)
	}

	return banners, nil
}

func (ur *UserRepo) GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanner")
	defer span.End()

	var banner entity.Banner

	err := ur.db.GetContext(spanCtx, &banner,
		`SELECT id, img_data, banner_text, banner_url, is_active, limit_shows,
       		limit_clicks, limit_budget, user_id, created_at, is_validated, comment, device, category_id
		FROM banners WHERE id=?`, bannerID)
	if err != nil {
		return nil, fmt.Errorf("could not get banner: %w", err)
	}

	return &banner, nil
}

const (
	tracerName = "db"
)

func (ur *UserRepo) CreateBanner(ctx context.Context, banner *entity.Banner) (int, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "CreateBanner")
	defer span.End()

	tx, err := ur.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("could not start tx: %w", err)
	}

	_, err = tx.ExecContext(spanCtx, `INSERT INTO banners (
                     	img_data, banner_text, banner_url, is_active, limit_shows, 
                     	limit_clicks, limit_budget, user_id, created_at, is_validated, comment, device, category_id)
					VALUES (?,?,?,?,?,?,?,?,?, ?, ?, ?, ?)`,
		banner.ImgData,
		banner.BannerText,
		banner.BannerURL,
		banner.IsActive,
		banner.LimitShows,
		banner.LimitClicks,
		banner.LimitBudget,
		banner.UserID,
		banner.CreatedAt,
		banner.IsValidated,
		"",
		banner.Device,
		banner.CategoryID,
	)

	if err != nil {
		return 0, fmt.Errorf("could not insert banner to mysql: %w", err)
	}

	var userID int
	if err := tx.GetContext(spanCtx, &userID, "SELECT LAST_INSERT_ID()"); err != nil {
		if err := tx.Rollback(); err != nil {
			return 0, fmt.Errorf("could not rollback tx: %w", err)
		}
		return 0, fmt.Errorf("could not get banner id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("could not commit: %w", err)
	}

	return userID, nil
}

func (ur *UserRepo) GetCategories(ctx context.Context) ([]*entity.WebsiteCategory, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetCategories")
	defer span.End()

	var categories []*entity.WebsiteCategory

	if err := ur.db.SelectContext(spanCtx, &categories, "SELECT id, name FROM category"); err != nil {
		return nil, fmt.Errorf("could not select categories: %w", err)
	}

	return categories, nil
}

func (ur *UserRepo) AddCategory(ctx context.Context, category *entity.WebsiteCategory) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddCategory")
	defer span.End()

	if _, err := ur.db.ExecContext(spanCtx, "INSERT INTO category (name) VALUES(?)", category.Name); err != nil {
		return fmt.Errorf("could not select categories: %w", err)
	}

	return nil
}

func (ur *UserRepo) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "WithTransaction")
	defer span.End()

	tx, err := ur.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	err = fn(injectTx(spanCtx, tx))
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

func (ur *UserRepo) BannerActivate(ctx context.Context, bannerID int, userID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerActivate")
	defer span.End()

	conn := ur.executor(spanCtx)

	res, err := conn.ExecContext(spanCtx, "UPDATE banners SET is_active=? WHERE id=? AND user_id=?", true, bannerID, userID)
	if err != nil {
		return fmt.Errorf("could not update: %w", err)
	}

	id, _ := res.RowsAffected()
	if id == 0 {
		return fmt.Errorf("update took no effect")
	}

	return nil
}

func (ur *UserRepo) BannerDeactivate(ctx context.Context, bannerID int, userID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerDeactivate")
	defer span.End()

	conn := ur.executor(spanCtx)

	res, err := conn.ExecContext(spanCtx, "UPDATE banners SET is_active=? WHERE id=? AND user_id=?", false, bannerID, userID)
	if err != nil {
		return fmt.Errorf("could not update: %w", err)
	}

	id, _ := res.RowsAffected()
	if id == 0 {
		return fmt.Errorf("update took no effect")
	}

	return nil
}

func (ur *UserRepo) BannerChangeStatus(ctx context.Context, bannerID int, status bool, comment string) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "BannerChangeStatus")
	defer span.End()

	// TODO: there are some business-logic that should be moved to services level
	// TODO: i gived up that logic for now, think about it later
	// If Admin rejects user's banner during its live then the banner needs to be stopped

	// type bannerForUpdate struct {
	// 	IsActive    bool `db:"is_active"`
	// 	IsValidated bool `db:"is_validated"`
	// }
	//
	// tx, err := ur.db.Beginx()
	// if err != nil {
	// 	return false, fmt.Errorf("could not start tx: %w", err)
	// }
	//
	// var banner bannerForUpdate
	// if err := tx.GetContext(ctx, &banner, "SELECT is_active, is_validated FROM banners WHERE id=? FOR UPDATE", bannerID); err != nil {
	// 	if err := tx.Rollback(); err != nil {
	// 		return false, fmt.Errorf("could not rollback tx: %w", err)
	// 	}
	//
	// 	return false, fmt.Errorf("could not get banner for update by ID: %w", err)
	// }
	//
	// var isStopped bool
	//
	// if banner.IsActive && !status {
	// 	isStopped = true
	// }
	//
	// _, err := ur.db.ExecContext(ctx, "UPDATE banners SET is_active=? WHERE id = ?", status, bannerID)
	// if err != nil {
	// 	return false, fmt.Errorf("could not update banner: %w", err)
	// }
	//
	// return nil

	_, err := ur.db.ExecContext(spanCtx, "UPDATE banners SET is_validated=?, comment=? WHERE id = ?", status, comment, bannerID)
	if err != nil {
		return fmt.Errorf("could not update banner: %w", err)
	}

	return nil
}
