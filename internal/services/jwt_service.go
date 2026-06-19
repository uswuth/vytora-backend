package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/user"
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

func (s *JWTService) GenerateToken(user *user.User) (string, error) {
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

func (s *JWTService) TokenTTL(tokenString string) (int64, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	if claims.ExpiresAt == nil {
		return 0, fmt.Errorf("token has no expiry")
	}
	return int64(time.Until(claims.ExpiresAt.Time).Seconds()), nil
}

func (s *JWTService) ExtendToken(tokenString string) (string, int64, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", 0, err
	}

	// Rebuild a minimal user from claims to generate a fresh token
	u := &user.User{
		ID:    uuid.MustParse(claims.UserID),
		Code:  claims.Code,
		Email: claims.Email,
		Role:  claims.Role,
	}

	newToken, err := s.GenerateToken(u)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	ttl, err := s.TokenTTL(newToken)
	if err != nil {
		return "", 0, err
	}

	return newToken, ttl, nil
}