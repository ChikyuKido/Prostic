package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
	jwtSecretErr  error
)

func CreateJWT() (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
	})

	return token.SignedString(secret)
}

func ValidateJWT(tokenString string) error {
	secret, err := getJWTSecret()
	if err != nil {
		return err
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}

		return secret, nil
	})
	if err != nil {
		return err
	}
	if !parsedToken.Valid {
		return errors.New("invalid token")
	}

	return nil
}

func getJWTSecret() ([]byte, error) {
	jwtSecretOnce.Do(func() {
		envSecret := os.Getenv("PROSTIC_JWT_SECRET")
		if envSecret != "" {
			jwtSecret = []byte(envSecret)
			return
		}

		rawSecret := make([]byte, 32)
		if _, err := rand.Read(rawSecret); err != nil {
			jwtSecretErr = err
			return
		}

		jwtSecret = []byte(base64.StdEncoding.EncodeToString(rawSecret))
	})

	return jwtSecret, jwtSecretErr
}
