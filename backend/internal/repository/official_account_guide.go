package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
	"yuanju/pkg/database"
)

const OfficialAccountGuideSettingKey = "miniapp_official_account_guide"
const OfficialAccountQRPrefix = "/static/uploads/official-account-qr/"

var officialAccountQRPath = regexp.MustCompile(`^/static/uploads/official-account-qr/[0-9a-fA-F-]{36}\.(png|jpg|webp)$`)

type OfficialAccountGuide struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	WechatID    string `json:"wechat_id"`
	Description string `json:"description"`
	QRPath      string `json:"qr_path"`
}

func (guide OfficialAccountGuide) Ready() bool {
	return guide.Enabled && strings.TrimSpace(guide.Name) != "" && strings.TrimSpace(guide.WechatID) != "" &&
		strings.TrimSpace(guide.Description) != "" && officialAccountQRPath.MatchString(guide.QRPath)
}

func ValidateOfficialAccountGuide(guide OfficialAccountGuide) (OfficialAccountGuide, error) {
	guide.Name = strings.TrimSpace(guide.Name)
	guide.WechatID = strings.TrimSpace(guide.WechatID)
	guide.Description = strings.TrimSpace(guide.Description)
	guide.QRPath = strings.TrimSpace(guide.QRPath)
	if utf8.RuneCountInString(guide.Name) > 40 || utf8.RuneCountInString(guide.WechatID) > 64 || utf8.RuneCountInString(guide.Description) > 120 {
		return guide, fmt.Errorf("公众号名称、微信号或说明过长")
	}
	if guide.QRPath != "" && !officialAccountQRPath.MatchString(guide.QRPath) {
		return guide, fmt.Errorf("二维码路径无效")
	}
	if guide.Enabled && !guide.Ready() {
		return guide, fmt.Errorf("启用前请填写公众号名称、微信号、说明并上传二维码")
	}
	return guide, nil
}

func GetOfficialAccountGuide() (OfficialAccountGuide, error) {
	var raw string
	err := database.DB.QueryRow(`SELECT value FROM system_settings WHERE key=$1`, OfficialAccountGuideSettingKey).Scan(&raw)
	if err == sql.ErrNoRows {
		return OfficialAccountGuide{}, nil
	}
	if err != nil {
		return OfficialAccountGuide{}, err
	}
	var guide OfficialAccountGuide
	if err := json.Unmarshal([]byte(raw), &guide); err != nil {
		return OfficialAccountGuide{}, fmt.Errorf("decode official account guide: %w", err)
	}
	return guide, nil
}

func SaveOfficialAccountGuide(guide OfficialAccountGuide) (OfficialAccountGuide, error) {
	guide, err := ValidateOfficialAccountGuide(guide)
	if err != nil {
		return guide, err
	}
	raw, err := json.Marshal(guide)
	if err != nil {
		return guide, err
	}
	_, err = database.DB.Exec(`INSERT INTO system_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()`, OfficialAccountGuideSettingKey, string(raw))
	return guide, err
}
