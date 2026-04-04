package middleware

import (
	"net/http"
	"strings"
	"yuanju/configs"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AdminAuth Admin 专属 JWT 鉴权中间件（issuer: yuanju-admin）
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供 Admin Token"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(configs.AppConfig.AdminJWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin Token 无效"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token 解析失败"})
			return
		}
		// 验证 issuer 必须是 yuanju-admin
		if iss, _ := claims["iss"].(string); iss != "yuanju-admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token 类型错误，请使用 Admin Token"})
			return
		}
		c.Set("admin_id", claims["sub"])
		c.Set("admin_email", claims["email"])
		c.Next()
	}
}
