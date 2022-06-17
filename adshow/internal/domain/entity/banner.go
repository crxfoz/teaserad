package entity

import "math/rand"

const (
	DeviceDesktop = "desktop"
	DeviceMobile  = "mobile"
	DeviceTablet  = "tablet"
)

type Banner struct {
	PlatformID int    `json:"platform_id"`
	BannerID   int    `json:"banner_id"`
	UserID     int    `json:"user_id"`
	ImgData    []byte `json:"img_data"`
	BannerText string `json:"banner_text"`
	BannerURL  string `json:"banner_url"`
	Device     string `json:"device"`
	CategoryID int    `json:"category_id"`
}

type Platform struct {
	PlatformID int `json:"platform_id" db:"platform_id"`
	CategoryID int `json:"category_id" db:"category_id"`
}

type Platforms []*Platform

func (p Platforms) Shuffle() {
	rand.Shuffle(len(p), func(i, j int) { p[i], p[j] = p[j], p[i] })
}

func (p Platforms) Ids() []int {
	out := make([]int, 0, len(p))

	for _, item := range p {
		out = append(out, item.PlatformID)
	}

	return out
}

type HitContext struct {
	UserAgent  string `json:"user_agent"`
	PlatformID int    `json:"platform_id"`
}

func (hc HitContext) DeviceType() string {
	// TODO: device type
	return DeviceDesktop
}
