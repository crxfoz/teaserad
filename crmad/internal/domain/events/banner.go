package events

type BannerUpdated struct {
	BannerID int    `json:"banner_id"`
	Valide   bool   `json:"valide"`
	Comment  string `json:"comment"`
}

type BannerCreated struct {
	BannerID   int    `json:"banner_id"`
	UserID     int    `json:"user_id"`
	Validated  bool   `json:"validated"`
	Device     string `json:"device" `
	CategoryID int    `json:"category_id"`
	CreatedAt  int64  `json:"created_at"`
}

type BannerStart struct {
	BannerID    int     `json:"banner_id"`
	UserID      int     `json:"user_id"`
	ImgData     []byte  `json:"img_data"`
	BannerText  string  `json:"banner_text"`
	BannerURL   string  `json:"banner_url"`
	LimitShows  int64   `json:"limit_shows"`
	LimitClicks int64   `json:"limit_clicks"`
	LimitBudget float64 `json:"limit_budget"`
	Device      string  `json:"device"`
	CategoryID  int     `json:"category_id"`
}

type BannerStop struct {
	BannerID int `json:"banner_id"`
}

type BannerReachedLimits struct {
	BannerID int    `json:"banner_id"`
	Reason   string `json:"reason"`
}
