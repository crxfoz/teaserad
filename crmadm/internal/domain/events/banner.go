package events

type BannerCreated struct {
	BannerID  int  `json:"banner_id"`
	UserID    int  `json:"user_id"`
	Validated bool `json:"validated"`
}

func (c *BannerCreated) ShouldBeAdded() bool {
	return !c.Validated
}

type BannerUpdated struct {
	BannerID int    `json:"banner_id"`
	Valide   bool   `json:"valide"`
	Comment  string `json:"comment"`
}
