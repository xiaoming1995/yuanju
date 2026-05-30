package service

import (
	"errors"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrRegistrationDisabled = errors.New("公开注册暂未开放")

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

type AdminCreateUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(input RegisterInput) (*model.User, string, error) {
	enabled, err := repository.GetBoolSetting(repository.SettingRegistrationEnabled, true)
	if err != nil {
		return nil, "", err
	}
	if !enabled {
		return nil, "", ErrRegistrationDisabled
	}

	user, err := createOrdinaryUser(input.Email, input.Password, input.Nickname, "self_registered")
	if err != nil {
		return nil, "", err
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func CreateUserByAdmin(input AdminCreateUserInput) (*model.User, error) {
	return createOrdinaryUser(input.Email, input.Password, input.Nickname, "admin_created")
}

func createOrdinaryUser(email, password, nickname, source string) (*model.User, error) {
	// 检查邮箱是否已存在
	existing, err := repository.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("该邮箱已被注册")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if nickname == "" {
		nickname = defaultNickname(email)
	}

	return repository.CreateUserWithSource(email, string(hash), nickname, source)
}

func defaultNickname(email string) string {
	local := strings.Split(email, "@")[0]
	if len([]rune(local)) > 5 {
		return string([]rune(local)[:5]) + "***"
	}
	if local != "" {
		return local + "***"
	}
	return "用户***"
}

func Login(input LoginInput) (*model.User, string, error) {
	user, err := repository.GetUserByEmail(input.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("邮箱或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, "", errors.New("邮箱或密码错误")
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func ResetUserPassword(userID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return repository.UpdateUserPassword(userID, string(hash))
}

func generateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(configs.AppConfig.JWTSecret))
}
