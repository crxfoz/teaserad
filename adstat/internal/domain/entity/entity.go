package entity

import "time"

type View struct {
	ViewID     string    `json:"view_id" db:"view_id"`
	BannerID   int       `json:"banner_id" db:"banner_id"`
	PlatformID int       `json:"platform_id" db:"platform_id"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	Device     string    `json:"device" db:"device"`
	CreatedAt  int64     `json:"created_at"`
	Day        string    `json:"day" db:"day"`
	DateTime   time.Time `json:"date_time" db:"dt"`
}

type ViewAggregated struct {
	BannerID   int    `json:"banner_id" db:"banner_id"`
	PlatformID int    `json:"platform_id" db:"platform_id"`
	Views      int    `json:"views" db:"hits"`
	Day        string `json:"day" db:"day"`
}

type Click struct {
	BannerID   int       `json:"banner_id" db:"banner_id"`
	PlatformID int       `json:"platform_id" db:"platform_id"`
	Price      float64   `json:"price" db:"price"`
	ViewID     string    `json:"view_id" db:"view_id"`
	CreatedAt  int64     `json:"created_at"`
	Day        string    `json:"day" db:"day"`
	DateTime   time.Time `json:"date_time" db:"dt"`
}

type ClickAggregated struct {
	BannerID   int     `json:"banner_id" db:"banner_id"`
	PlatformID int     `json:"platform_id" db:"platform_id"`
	Clicks     int     `json:"clicks" db:"clicks"`
	Price      float64 `json:"price" db:"price"`
	Day        string  `json:"day" db:"day"`
}

type BannerStat struct {
	BannerID int     `json:"banner_id" db:"banner_id"`
	Views    int     `json:"views" db:"hits"`
	Clicks   int     `json:"clicks" db:"clicks"`
	CTR      float64 `json:"ctr" db:"ctr"`
	Day      string  `json:"day" db:"day"`
}

type PlatformStat struct {
	PlatformID int     `json:"platform_id" db:"platform_id"`
	Views      int     `json:"views" db:"hits"`
	Clicks     int     `json:"clicks" db:"clicks"`
	CTR        float64 `json:"ctr" db:"ctr"`
	Day        string  `json:"day" db:"day"`
}
