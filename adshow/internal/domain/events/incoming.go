package events

type BannerStart struct {
	BannerID   int    `json:"banner_id"`
	UserID     int    `json:"user_id"`
	ImgData    []byte `json:"img_data"`
	BannerText string `json:"banner_text"`
	BannerURL  string `json:"banner_url"`
	// LimitShows  int64   `json:"limit_shows"`
	// LimitClicks int64   `json:"limit_clicks"`
	// LimitBudget float64 `json:"limit_budget"`
	Device     string `json:"device"`
	CategoryID int    `json:"category_id"`
}

type BannerStop struct {
	BannerID int `json:"banner_id"`
}
