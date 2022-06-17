package events

type View struct {
	BannerID   int    `json:"banner_id"`
	PlatformID int    `json:"platform_id"`
	UserAgent  string `json:"user_agent"`
	Device     string `json:"device"`
	CreatedAt  int64  `json:"created_at"`
}
