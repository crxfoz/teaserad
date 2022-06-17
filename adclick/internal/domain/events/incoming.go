package events

type NewBanner struct {
	BannerID  int    `json:"banner_id"`
	BannerURL string `json:"banner_url"`
}
