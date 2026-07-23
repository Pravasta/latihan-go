package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
	}
}

func (s *JWTService) Generate(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) Parse(token string) (string, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})

	if err != nil {
		fmt.Printf("[JWT] Failed to parse token: %v\n", err)
		return "", err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			fmt.Printf("[JWT] Invalid user_id claim\n")
			return "", jwt.ErrInvalidKey
		}
		return userID, nil
	}

	return "", jwt.ErrInvalidKey
}
