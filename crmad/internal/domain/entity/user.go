package entity

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	DeviceDesktop = "desktop"
	DeviceMobile  = "mobile"
	DeviceTablet  = "tablet"
)

type User struct {
	ID        int    `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"password" db:"password"`
	Validated bool   `json:"validated" db:"validated"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
}

func (u *User) CheckPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return fmt.Errorf("invalide password: %w", err)
	}

	return nil
}

func (u *User) HashedPassword() string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

type Banner struct {
	ID          int     `json:"id" db:"id"`
	UserID      int     `json:"user_id" db:"user_id"`
	ImgData     []byte  `json:"-" db:"img_data"`
	BannerText  string  `json:"banner_text" db:"banner_text"`
	BannerURL   string  `json:"banner_url" db:"banner_url"`
	IsActive    bool    `json:"is_active" db:"is_active"`
	LimitShows  int64   `json:"limit_shows" db:"limit_shows"`
	LimitClicks int64   `json:"limit_clicks" db:"limit_clicks"`
	LimitBudget float64 `json:"limit_budget" db:"limit_budget"`
	CreatedAt   int64   `json:"created_at" db:"created_at"`
	IsValidated bool    `json:"is_validated" db:"is_validated"`
	Comment     string  `json:"comment" db:"comment"`
	Device      string  `json:"device" db:"device"`
	CategoryID  int     `json:"category_id" db:"category_id"`
}

func (b *Banner) CanBeStarted() bool {
	return !b.IsActive && b.IsValidated
}

func (b *Banner) CanBeStopped() bool {
	return b.IsActive
}

func (b *Banner) Validate(categories []*WebsiteCategory) error {
	switch b.Device {
	case DeviceDesktop, DeviceTablet, DeviceMobile:
	default:
		return fmt.Errorf("wrong device: %s", b.Device)
	}

	for _, item := range categories {
		if b.CategoryID == item.ID {
			return nil
		}
	}

	return fmt.Errorf("invalide category: %d", b.CategoryID)
}

type Invoice struct {
	ID        int     `json:"id" db:"id"`
	Amount    float64 `json:"amount" db:"amount"`
	Key       string  `json:"key" db:"key"`
	IsPaid    bool    `json:"is_paid" db:"is_paid"`
	CreatedAt int64   `json:"created_at" db:"created_at"`
}

func (i *Invoice) GetAmount() float64 {
	if !i.IsPaid {
		return 0
	}

	return i.Amount
}
