package service

import (
	"errors"
	"time"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(input RegisterInput) (*model.User, string, error) {
	// 检查邮箱是否已存在
	existing, err := repository.GetUserByEmail(input.Email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("该邮箱已被注册")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	nickname := input.Nickname
	if nickname == "" {
		nickname = input.Email[:5] + "***"
	}

	user, err := repository.CreateUser(input.Email, string(hash), nickname)
	if err != nil {
		return nil, "", err
	}

	token, err := generateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
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

func generateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(configs.AppConfig.JWTSecret))
}
