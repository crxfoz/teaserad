package entity

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

type User struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"password" db:"password"`
	Role      string `json:"role" db:"role"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

func (u *User) CheckPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return fmt.Errorf("invalide password: %w", err)
	}

	return nil
}

type Banner struct {
	BannerID  int   `json:"banner_id" db:"banner_id"`
	UserID    int   `json:"user_id" db:"user_id"`
	CreatedAt int64 `json:"created_at" db:"created_at"`
}

type Resolution struct {
	BannerID  int    `json:"banner_id" db:"banner_id"`
	Valide    bool   `json:"valide" db:"valide"`
	Comment   string `json:"comment" db:"comment"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

type UserContext struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type BannerResulution struct {
	Banner
	Valide    bool   `json:"valide" db:"valide"`
	Comment   string `json:"comment" db:"comment"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}
