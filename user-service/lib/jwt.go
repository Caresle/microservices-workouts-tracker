package lib

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/caresle/microservices-workouts-tracker/user-service/models"
	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	UserId int    `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func getSecretKey() (string, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if strings.TrimSpace(secretKey) == "" {
		return "", errors.New("INVALID JWT SECRET")
	}

	return secretKey, nil
}

func GenerateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey, err := getSecretKey()

	if err != nil {
		return "", err
	}

	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(token string) (*AuthClaims, error) {
	secretKey, err := getSecretKey()

	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.ParseWithClaims(token, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*AuthClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("INVALID JWT TOKEN")
}
