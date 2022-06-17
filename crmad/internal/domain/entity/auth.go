package entity

type UserContext struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Validated bool   `json:"validated"`
	Token     string `json:"token"`
}
