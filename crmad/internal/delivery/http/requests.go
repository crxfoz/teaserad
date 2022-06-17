package http

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Register struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NewBanner struct {
	ImgData     string  `json:"img_data"`
	BannerText  string  `json:"banner_text"`
	BannerURL   string  `json:"banner_url"`
	LimitShows  int64   `json:"limit_shows"`
	LimitClicks int64   `json:"limit_clicks"`
	LimitBudget float64 `json:"limit_budget"`
	Device      string  `json:"device"`
	CategoryID  int     `json:"category_id"`
}
