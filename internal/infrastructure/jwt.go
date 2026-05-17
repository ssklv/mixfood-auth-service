package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenProvider struct {
	signingKey []byte
	accessTTL  time.Duration
}

//GenerateAccessToken(userID int64, role string) (string, error)
//GenerateRefreshToken() (string, error)
//ParseToken(tokenString string) (int64, string, error)

func NewTokenProvider(key string, ttl time.Duration) *TokenProvider {
	return &TokenProvider{
		signingKey: []byte(key),
		accessTTL:  ttl,
	}
}

func (p *TokenProvider) GenerateAccessToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(p.accessTTL).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.signingKey)
}

func (p *TokenProvider) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (p *TokenProvider) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return p.signingKey, nil
	})
	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["sub"].(float64))
		role, _ := claims["role"].(string)
		return userID, role, nil
	}
	return 0, "", errors.New("invalid token claims")
}
