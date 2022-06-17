package mysql

import (
	"context"
	"fmt"

	"github.com/crxfoz/teaserad/adshow/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

type PlatformRepo struct {
	conn *sqlx.DB
}

func New(conn *sqlx.DB) *PlatformRepo {
	return &PlatformRepo{conn: conn}
}

const (
	tracerName = "db-mysql"
)

// TODO: if crmweb add new platform we need to deliver ad there

func (r *PlatformRepo) GetPlatforms(ctx context.Context) (entity.Platforms, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetPlatforms")
	defer span.End()

	var platforms []*entity.Platform

	err := r.conn.SelectContext(spanCtx, &platforms, "SELECT platform_id, category_id FROM platforms")
	if err != nil {
		return nil, fmt.Errorf("could not get platforms: %w", err)
	}

	return platforms, nil
}

func (r *PlatformRepo) GetPlatformsByCategory(ctx context.Context, categoryID int) (entity.Platforms, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetPlatformsByCategory")
	defer span.End()

	var platforms []*entity.Platform

	err := r.conn.SelectContext(spanCtx, &platforms, "SELECT platform_id, category_id FROM platforms WHERE category_id=?", categoryID)
	if err != nil {
		return nil, fmt.Errorf("could not get platforms: %w", err)
	}

	return platforms, nil
}

func (r *PlatformRepo) GetPlatform(ctx context.Context, platformID int) (*entity.Platform, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetPlatform")
	defer span.End()

	var platforms entity.Platform

	err := r.conn.GetContext(spanCtx, &platforms, "SELECT platform_id, category_id FROM platforms WHERE platform_id=?", platformID)
	if err != nil {
		return nil, fmt.Errorf("could not get platform: %w", err)
	}

	return &platforms, nil
}

func (r *PlatformRepo) AddPlatform(ctx context.Context, platform *entity.Platform) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddPlatform")
	defer span.End()

	_, err := r.conn.ExecContext(spanCtx, "INSERT INTO platforms (platform_id, category_id) VALUES (?,?)", platform.PlatformID, platform.CategoryID)
	if err != nil {
		return fmt.Errorf("could not insert new platform: %w", err)
	}

	return nil
}

func (r *PlatformRepo) DeletePlatform(ctx context.Context, platformID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "DeletePlatform")
	defer span.End()

	rows, err := r.conn.ExecContext(spanCtx, "DELETE FROM platforms WHERE platform_id=?", platformID)
	if err != nil {
		return fmt.Errorf("could not delete platform: %w", err)
	}

	n, _ := rows.RowsAffected()
	if n == 0 {
		return fmt.Errorf("there is nothing to delete")
	}

	return nil
}
