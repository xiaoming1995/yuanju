package repository

import "testing"

func TestOfficialAccountGuideValidation(t *testing.T) {
	complete := OfficialAccountGuide{
		Enabled: true, Name: " 缘聚 ", WechatID: " yuanju ", Description: " 内容更新 ",
		QRPath: "/static/uploads/official-account-qr/550e8400-e29b-41d4-a716-446655440000.png",
	}
	got, err := ValidateOfficialAccountGuide(complete)
	if err != nil {
		t.Fatalf("valid guide rejected: %v", err)
	}
	if got.Name != "缘聚" || got.WechatID != "yuanju" || got.Description != "内容更新" {
		t.Fatalf("fields not trimmed: %+v", got)
	}
	if !got.Ready() {
		t.Fatal("complete guide is not ready")
	}

	missing := complete
	missing.WechatID = ""
	if _, err := ValidateOfficialAccountGuide(missing); err == nil {
		t.Fatal("enabled guide without search ID accepted")
	}
	missing.Enabled = false
	if _, err := ValidateOfficialAccountGuide(missing); err != nil {
		t.Fatalf("disabled draft rejected: %v", err)
	}
	if missing.Ready() {
		t.Fatal("disabled guide must not be public")
	}

	external := complete
	external.QRPath = "https://other.example/qr.png"
	if _, err := ValidateOfficialAccountGuide(external); err == nil {
		t.Fatal("external URL accepted")
	}
	traversal := complete
	traversal.QRPath = "/static/uploads/official-account-qr/../user-avatars/abc.png"
	if _, err := ValidateOfficialAccountGuide(traversal); err == nil {
		t.Fatal("path traversal accepted")
	}
}

func TestOfficialAccountGuideReadyRejectsWhitespaceFromStoredData(t *testing.T) {
	guide := OfficialAccountGuide{
		Enabled: true, Name: "  ", WechatID: "yuanju", Description: "内容更新",
		QRPath: "/static/uploads/official-account-qr/550e8400-e29b-41d4-a716-446655440000.png",
	}
	if guide.Ready() {
		t.Fatal("whitespace-only name must not be exposed")
	}
}

func TestOfficialAccountGuideStorage(t *testing.T) {
	db := withArticleRepositoryDB(t)
	guide, err := GetOfficialAccountGuide()
	if err != nil || guide.Enabled {
		t.Fatalf("default: %+v %v", guide, err)
	}
	guide = OfficialAccountGuide{Enabled: true, Name: "缘聚", WechatID: "yuanju", Description: "内容更新",
		QRPath: "/static/uploads/official-account-qr/550e8400-e29b-41d4-a716-446655440000.png"}
	if _, err := SaveOfficialAccountGuide(guide); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := GetOfficialAccountGuide()
	if err != nil || got != guide {
		t.Fatalf("read: %+v %v", got, err)
	}
	if _, err := db.Exec(`UPDATE system_settings SET value='{broken' WHERE key=$1`, OfficialAccountGuideSettingKey); err != nil {
		t.Fatal(err)
	}
	if _, err := GetOfficialAccountGuide(); err == nil {
		t.Fatal("broken JSON should return error")
	}
	if _, err := db.Exec(`DROP TABLE system_settings`); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveOfficialAccountGuide(guide); err == nil {
		t.Fatal("database write failure should propagate")
	}
}
