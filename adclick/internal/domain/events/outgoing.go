package events

type Click struct {
	BannerID   int     `json:"banner_id"`
	PlatformID int     `json:"platform_id"`
	ViewID     string  `json:"view_id"`
	Price      float64 `json:"price"`
	CreatedAt  int64   `json:"created_at"`
}
