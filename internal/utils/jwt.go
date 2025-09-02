package utils

import (
	"errors"
	"time"
	"xquant-default-management/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims 是我们存储在 JWT 中的自定义数据
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 为指定用户生成一个新的 JWT
func GenerateToken(userID uuid.UUID, role string, cfg config.Config) (string, error) {
	// 设置 token 的过期时间
	expirationTime := time.Now().Add(time.Duration(cfg.TokenTTL) * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "xquant-default-management",
		},
	}

	// 使用 HS256 签名算法创建一个新的 token 对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用我们在配置中定义的 secret 来签名 token
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证给定的 token 字符串
func ValidateToken(tokenString string, cfg config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
