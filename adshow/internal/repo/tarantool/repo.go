package tarantool

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/crxfoz/teaserad/adshow/internal/domain"
	"github.com/crxfoz/teaserad/adshow/internal/domain/entity"
	"github.com/crxfoz/teaserad/adshow/internal/domain/events"
	"github.com/tarantool/go-tarantool"
	"go.opentelemetry.io/otel"
)

// TODO: this storage has bad consistency and the whole design

type Cluster interface {
	Node(int) *tarantool.Connection
	Nodes() []*tarantool.Connection
}

type ShowRepo struct {
	cluster Cluster
	logger  domain.Logger
}

const (
	tracerName = "db-tarantool"
)

func New(cluster Cluster, logger domain.Logger) *ShowRepo {
	rand.Seed(time.Now().Unix())
	return &ShowRepo{cluster: cluster, logger: logger}
}

func (r *ShowRepo) AddBanner(ctx context.Context, start *events.BannerStart, toPlatforms []int) error {
	_, span := otel.Tracer(tracerName).Start(ctx, "AddBanner")
	defer span.End()

	for _, platformID := range toPlatforms {
		conn := r.cluster.Node(platformID)
		_, err := conn.Insert("platforms", []interface{}{
			platformID,
			start.Device,
			start.BannerID,
			start.BannerURL,
			start.BannerText,
			start.CategoryID,
			start.ImgData,
			start.UserID})
		if err != nil {
			r.logger.Errorw("could not insert banner", "err", err, "platformID", platformID)
		}

	}

	return nil
}

func (r *ShowRepo) scanTuple(tuple []interface{}) (*entity.Banner, error) {
	if len(tuple) != 8 {
		return nil, fmt.Errorf("wrong len: %d", len(tuple))
	}

	banner := new(entity.Banner)

	pid, ok := tuple[0].(uint64)
	if !ok {
		return nil, fmt.Errorf("could not parse PlatformID")
	}
	banner.PlatformID = int(pid)

	dev, ok := tuple[1].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse Device")
	}
	banner.Device = dev

	bid, ok := tuple[2].(uint64)
	if !ok {
		return nil, fmt.Errorf("could not parse BannerID")
	}
	banner.BannerID = int(bid)

	uri, ok := tuple[3].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse BannerURL")
	}
	banner.BannerURL = uri

	btext, ok := tuple[4].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse BannerText")
	}
	banner.BannerText = btext

	cid, ok := tuple[5].(uint64)
	if !ok {
		return nil, fmt.Errorf("could not parse CategoryID")
	}
	banner.CategoryID = int(cid)

	img, ok := tuple[6].([]byte)
	if !ok {
		return nil, fmt.Errorf("could not parse ImgData")
	}
	banner.ImgData = img

	uid, ok := tuple[7].(uint64)
	if !ok {
		return nil, fmt.Errorf("could not parse UserID")
	}
	banner.UserID = int(uid)

	return banner, nil
}

func (r *ShowRepo) bannerToPrimaryIndex(banner *entity.Banner) []interface{} {
	return []interface{}{banner.PlatformID, banner.Device, banner.BannerID}
}

func (r *ShowRepo) deleteBannerAll(ctx context.Context, conn *tarantool.Connection, bannerID int) (errRet error) {
	for limit, offset := 1000, 0; ; limit, offset = limit+1000, offset+1000 {
		resp, err := conn.Select("platforms", "secondary", uint32(offset), uint32(limit), tarantool.IterEq, []interface{}{bannerID})
		if err != nil {
			return fmt.Errorf("could not select platforms: %w", err)
		}

		if len(resp.Data) == 0 {
			break
		}

		var data []*entity.Banner
		for _, item := range resp.Tuples() {
			if v, err := r.scanTuple(item); err == nil {
				data = append(data, v)
			}
		}

		for _, bn := range data {
			if _, err := conn.Delete("platforms", "primary", r.bannerToPrimaryIndex(bn)); err != nil {
				return fmt.Errorf("could not delete")
			}
		}
	}

	return nil
}

func (r *ShowRepo) DeleteBannerAll(ctx context.Context, bannerID int) error {
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, "DeleteBannerAll")
	defer span.End()

	for _, node := range r.cluster.Nodes() {
		if err := r.deleteBannerAll(spanCtx, node, bannerID); err != nil {
			r.logger.Errorw("could not delete banner", "err", err)
		}
	}

	return nil
}

func (r *ShowRepo) selectLen(actualLen int, desiredLen int) int {
	if desiredLen > actualLen {
		return actualLen
	}

	return desiredLen
}

func (r *ShowRepo) BannersForPlatform(ctx context.Context, platformID int, deviceType string, limit int) ([]*entity.Banner, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "BannersForPlatform")
	defer span.End()

	conn := r.cluster.Node(platformID)
	var data []*entity.Banner

	for limit, offset := 1000, 0; ; limit, offset = limit+1000, offset+1000 {
		resp, err := conn.Select("platforms", "primary", uint32(offset), uint32(limit), tarantool.IterEq, []interface{}{platformID, deviceType})
		if err != nil {
			return nil, fmt.Errorf("could not select platforms: %w", err)
		}

		if len(resp.Data) == 0 {
			break
		}

		for _, item := range resp.Tuples() {
			if v, err := r.scanTuple(item); err == nil {
				data = append(data, v)
			}
		}
	}

	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	return data[:r.selectLen(len(data), limit)], nil
}
