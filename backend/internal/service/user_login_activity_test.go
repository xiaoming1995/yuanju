package service

import (
	"context"
	"testing"

	"yuanju/internal/repository"
)

func TestWebLoginActivityDistinguishesRegistrationAndFailedLogin(t *testing.T) {
	withAuthTestDB(t)
	if err := repository.SetBoolSetting(repository.SettingRegistrationEnabled, true); err != nil {
		t.Fatal(err)
	}
	user, _, err := Register(RegisterInput{Email: "activity-login@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	events, total, err := repository.ListUserActivity(user.ID, 1, 20)
	if err != nil || total != 1 || events[0].Type != "register" {
		t.Fatalf("registration events=%+v total=%d err=%v", events, total, err)
	}
	detail, err := repository.GetAdminUserDetail(user.ID)
	if err != nil || detail.LastLoginAt != nil {
		t.Fatalf("registration counted as login: %+v %v", detail, err)
	}
	if _, _, err := LoginForWeb(context.Background(), LoginInput{Email: user.Email, Password: "wrong"}); err == nil {
		t.Fatal("wrong password succeeded")
	}
	events, total, err = repository.ListUserActivity(user.ID, 1, 20)
	if err != nil || total != 1 {
		t.Fatalf("failed login events=%+v total=%d err=%v", events, total, err)
	}
	if _, _, err := LoginForWeb(context.Background(), LoginInput{Email: user.Email, Password: "password123"}); err != nil {
		t.Fatal(err)
	}
	events, total, err = repository.ListUserActivity(user.ID, 1, 20)
	if err != nil || total != 2 || events[0].Type != "login" || events[0].Channel != "web" {
		t.Fatalf("login events=%+v total=%d err=%v", events, total, err)
	}
	detail, err = repository.GetAdminUserDetail(user.ID)
	if err != nil || detail.LastLoginAt == nil || detail.LastActiveAt == nil || detail.LastLoginChannel != "web" {
		t.Fatalf("last login: %+v %v", detail, err)
	}
}
