package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func spinUpAuthPG(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("yuanju_test"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("conn string: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("ping db: %v", err)
	}

	return db, func() {
		_ = db.Close()
		_ = pg.Terminate(ctx)
	}
}

func withAuthTestDB(t *testing.T) {
	t.Helper()
	db, cleanup := spinUpAuthPG(t)
	t.Cleanup(cleanup)

	origDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = origDB })

	configs.AppConfig.JWTSecret = "test-user-secret"
	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestRegisterRespectsPublicRegistrationSetting(t *testing.T) {
	withAuthTestDB(t)

	if err := repository.SetBoolSetting(repository.SettingRegistrationEnabled, false); err != nil {
		t.Fatalf("disable registration: %v", err)
	}

	user, token, err := Register(RegisterInput{
		Email:    "blocked@example.com",
		Password: "password123",
		Nickname: "blocked",
	})
	if !errors.Is(err, ErrRegistrationDisabled) {
		t.Fatalf("expected ErrRegistrationDisabled, got user=%v token=%q err=%v", user, token, err)
	}
	if existing, err := repository.GetUserByEmail("blocked@example.com"); err != nil || existing != nil {
		t.Fatalf("disabled registration should not create user, existing=%v err=%v", existing, err)
	}

	if err := repository.SetBoolSetting(repository.SettingRegistrationEnabled, true); err != nil {
		t.Fatalf("enable registration: %v", err)
	}
	user, token, err = Register(RegisterInput{
		Email:    "open@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("enabled registration should succeed: %v", err)
	}
	if user == nil || token == "" {
		t.Fatalf("expected created user and token, got user=%v token=%q", user, token)
	}
	if user.Source != "self_registered" {
		t.Fatalf("source=%q, want self_registered", user.Source)
	}
}

func TestAdminCreatedUserBypassesPublicRegistrationSettingAndCanLogin(t *testing.T) {
	withAuthTestDB(t)

	if err := repository.SetBoolSetting(repository.SettingRegistrationEnabled, false); err != nil {
		t.Fatalf("disable registration: %v", err)
	}

	user, err := CreateUserByAdmin(AdminCreateUserInput{
		Email:    "admin-created@example.com",
		Password: "password123",
		Nickname: "后台用户",
	})
	if err != nil {
		t.Fatalf("admin create user: %v", err)
	}
	if user.Source != "admin_created" {
		t.Fatalf("source=%q, want admin_created", user.Source)
	}

	loggedIn, token, err := Login(LoginInput{
		Email:    "admin-created@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("admin-created user login: %v", err)
	}
	if loggedIn == nil || token == "" {
		t.Fatalf("expected logged in user and token, got user=%v token=%q", loggedIn, token)
	}
}
