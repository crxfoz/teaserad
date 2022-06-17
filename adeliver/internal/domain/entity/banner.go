package entity

type Banner struct {
	ID          int     `json:"id"`
	LimitShows  int64   `json:"limit_shows"`
	LimitClicks int64   `json:"limit_clicks"`
	LimitBudget float64 `json:"limit_budget"`
}
