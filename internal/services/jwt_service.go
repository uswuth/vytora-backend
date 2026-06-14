package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uswuth/vytora-backend/internal/models"
)

type JWTService struct {
	secret      string
	expiryHours int
}

func NewJWTService(secret string, expiryHours int) *JWTService {
	return &JWTService{secret: secret, expiryHours: expiryHours}
}

type Claims struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID.String(),
		Code:   user.Code,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Code,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
