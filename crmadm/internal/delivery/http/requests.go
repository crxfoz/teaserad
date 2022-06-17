package http

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResolutinRequest struct {
	BannerID int    `json:"banner_id" db:"banner_id"`
	Valide   bool   `json:"valide" db:"valide"`
	Comment  string `json:"comment" db:"comment"`
}
