package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/crxfoz/teaserad/adstat/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	conn *sqlx.DB
}

func New(conn *sqlx.DB) *Repo {
	return &Repo{conn: conn}
}

func (r *Repo) clickhouseDayFormat(in time.Time) string {
	return in.Format("2006-01-02")
}

func (r *Repo) GetBannerStat(ctx context.Context, bannerID int, from time.Time) ([]*entity.BannerStat, error) {
	tm := r.clickhouseDayFormat(from)
	var stat []*entity.BannerStat

	err := r.conn.SelectContext(ctx, &stat, "SELECT h.banner_id,sum(c.clicks)clicks,sum(h.hits)hits,h.day,(clicks/hits*100)AS ctr FROM hits_daily h LEFT JOIN(SELECT c.banner_id,sum(c.clicks)AS clicks,c.day FROM clicks_daily c GROUP BY c.day,c.banner_id)c ON h.banner_id=c.banner_id AND h.day=c.day WHERE h.banner_id=? AND h.day>= ? GROUP BY h.day,h.banner_id",
		bannerID,
		tm)
	if err != nil {
		return nil, fmt.Errorf("could not select: %w", err)
	}

	return stat, nil
}

func (r *Repo) GetPlatformStat(ctx context.Context, platformID int, from time.Time) ([]*entity.PlatformStat, error) {
	tm := r.clickhouseDayFormat(from)
	var stat []*entity.PlatformStat

	err := r.conn.SelectContext(ctx, &stat, "SELECT h.platform_id,sum(c.clicks)clicks,sum(h.hits)hits,h.day,(clicks/hits*100)AS ctr FROM hits_daily h LEFT JOIN(SELECT c.platform_id,sum(c.clicks)AS clicks,c.day FROM clicks_daily c GROUP BY c.day,c.platform_id)c ON h.platform_id=c.platform_id AND h.day=c.day WHERE h.platform_id=? AND h.day>= ? GROUP BY h.day,h.platform_id",
		platformID,
		tm)
	if err != nil {
		return nil, fmt.Errorf("could not select: %w", err)
	}

	return stat, nil
}
