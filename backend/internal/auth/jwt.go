package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultSecret = "dvr-manager-secret-key-change-in-production"

// Claims JWT 载荷
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWT 签发与校验
type JWT struct {
	secret []byte
}

// NewJWT 从环境变量或参数创建 JWT 服务
func NewJWT(secret string) *JWT {
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		secret = defaultSecret
	}
	return &JWT{secret: []byte(secret)}
}

// Generate 签发 Token（24h）
func (j *JWT) Generate(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// Verify 校验 Token
func (j *JWT) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ExtractBearer 从 Authorization 头提取 Bearer Token
func ExtractBearer(header string) string {
	header = strings.TrimSpace(header)
	if len(header) > 7 && strings.EqualFold(header[:7], "Bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return header
}
