package utils

import (
	"errors"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var JWTSecret = []byte(os.Getenv("JWT_SECRET")) // need to update

func GenerateJWT(userId uint, email string) (string, error) {
	claims := Claims{
		UserID: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ValidateFile(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > 100*1024 {
		return errors.New("file size exceeds 100KiB")
	}

	ext := strings.ToLower(strings.TrimPrefix(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):], "."))
	if ext != "jpeg" && ext != "jpg" && ext != "png" {
		return errors.New("invalid file type; only jpeg, jpg, png allowed")
	}
	return nil
}
