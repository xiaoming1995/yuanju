package model

import "time"

// ExportBrand represents a user's customization of export PNG/PDF branding.
type ExportBrand struct {
	UserID        string    `json:"-"`
	Title         string    `json:"title"`
	FooterText    string    `json:"footer_text"`
	LogoPath      string    `json:"-"`
	LogoURL       string    `json:"logo_url"`
	WatermarkMode string    `json:"watermark_mode"`
	WatermarkText string    `json:"watermark_text"`
	UpdatedAt     time.Time `json:"-"`
}
