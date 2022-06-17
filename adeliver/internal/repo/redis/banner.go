package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/crxfoz/teaserad/adeliver/internal/domain/entity"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
)

const (
	fieldClick = "click"
	fieldShow  = "show"
	fieldSpend = "spend"
)

type Cluster interface {
	Node(int) *redis.Client
}

type Redis struct {
	cluster Cluster
}

func New(cluster Cluster) *Redis {
	return &Redis{
		cluster: cluster,
	}
}

const (
	tracerName = "db-redis"
)

func (r *Redis) ketInteractions(bannerID int, kind string) string {
	return fmt.Sprintf("interactions.%d.%s", bannerID, kind)
}

func (r *Redis) keyInfo(bannerID int) string {
	return fmt.Sprintf("info.%d", bannerID)
}

func (r *Redis) GetBanner(ctx context.Context, bannerID int) (*entity.Banner, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetBanner")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Get(spanCtx, r.keyInfo(bannerID))
	err := res.Err()
	if err != nil {
		return nil, fmt.Errorf("could nmot store info: %w", err)
	}

	out, err := res.Result()
	if err != nil {
		return nil, fmt.Errorf("could not extract result: %w", err)
	}

	var banner entity.Banner
	if err := json.Unmarshal([]byte(out), &banner); err != nil {
		return nil, fmt.Errorf("could not unmarshal: %w", err)
	}

	return &banner, nil
}

func (r *Redis) AddBanner(ctx context.Context, banner entity.Banner) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddBanner")
	defer span.End()

	conn := r.cluster.Node(banner.ID)

	raw, err := json.Marshal(banner)
	if err != nil {
		return fmt.Errorf("could not marshal: %w", err)
	}

	res := conn.Set(spanCtx, r.keyInfo(banner.ID), raw, 0)
	err = res.Err()
	if err != nil {
		return fmt.Errorf("could not store info: %w", err)
	}

	return nil
}

func (r *Redis) GetClick(ctx context.Context, bannerID int) (int64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetClick")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Get(spanCtx, r.ketInteractions(bannerID, fieldClick))
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not get clicks: %w", err)
	}

	outStr, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	out, err := strconv.ParseInt(outStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not cast string: %w", err)
	}

	return out, nil
}

func (r *Redis) GetShows(ctx context.Context, bannerID int) (int64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetShows")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Get(spanCtx, r.ketInteractions(bannerID, fieldShow))
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not get shows: %w", err)
	}

	outStr, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	out, err := strconv.ParseInt(outStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not cast string: %w", err)
	}

	return out, nil
}

func (r *Redis) GetSpend(ctx context.Context, bannerID int) (float64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "GetSpend")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Get(spanCtx, r.ketInteractions(bannerID, fieldSpend))
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not get spend: %w", err)
	}

	outStr, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	out, err := strconv.ParseFloat(outStr, 64)
	if err != nil {
		return 0, fmt.Errorf("could not cast string: %w", err)
	}

	return out, nil
}

func (r *Redis) AddClick(ctx context.Context, bannerID int) (int64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddClick")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Incr(spanCtx, r.ketInteractions(bannerID, fieldClick))
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not incr clicks: %w", err)
	}

	out, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	return out, nil
}

func (r *Redis) AddShow(ctx context.Context, bannerID int) (int64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddShow")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.Incr(spanCtx, r.ketInteractions(bannerID, fieldShow))
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not incr shows: %w", err)
	}

	out, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	return out, nil
}

func (r *Redis) AddSpend(ctx context.Context, bannerID int, price float64) (float64, error) {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "AddSpend")
	defer span.End()

	conn := r.cluster.Node(bannerID)

	res := conn.IncrByFloat(spanCtx, r.ketInteractions(bannerID, fieldShow), price)
	err := res.Err()
	if err != nil {
		return 0, fmt.Errorf("could not incr shows: %w", err)
	}

	out, err := res.Result()
	if err != nil {
		return 0, fmt.Errorf("could not get result: %w", err)
	}

	return out, nil
}
