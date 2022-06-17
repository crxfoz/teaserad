package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crxfoz/teaserad/adclick/internal/domain/entity"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
)

type Cluster interface {
	Node(int) *redis.Client
}

type Redis struct {
	cluster Cluster
}

func New(cluster Cluster) *Redis {
	return &Redis{cluster: cluster}
}

func (r *Redis) buckedByID(bannerID int) string {
	return fmt.Sprintf("banner.url.%d", bannerID)
}

const (
	tracerName = "db-redis"
)

func (r *Redis) AddBanner(ctx context.Context, banner *entity.BannerURL) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddBanner")
	defer span.End()

	conn := r.cluster.Node(banner.BannerID)

	out, err := json.Marshal(banner)
	if err != nil {
		return fmt.Errorf("could not marshal: %w", err)
	}

	cmd := conn.Set(spanCtx, r.buckedByID(banner.BannerID), out, 0)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("could not add banner: %w", err)
	}

	return nil
}

func (r *Redis) GetBanner(ctx context.Context, bannerID int) (*entity.BannerURL, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanner")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	cmd := conn.Get(spanCtx, r.buckedByID(bannerID))
	if err := cmd.Err(); err != nil {
		return nil, fmt.Errorf("could not get data: %w", err)
	}

	out, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("could not extract result: %w", err)
	}

	var banner entity.BannerURL
	if err := json.Unmarshal([]byte(out), &banner); err != nil {
		return nil, fmt.Errorf("could not unmarshal: %w", err)
	}

	return &banner, nil
}
