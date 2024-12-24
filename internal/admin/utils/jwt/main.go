package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pingcap/log"
	"os"
	"time"
)

func CreateTokenForAdmin(adminUsername string) (string, error) {
	hmacSecret := []byte(os.Getenv("SECRET_KEY"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_username": adminUsername,
		"exp":            time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	hmacSecret := []byte(os.Getenv("SECRET_KEY"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSecret, nil
	})
	if err != nil {
		log.Error("Error in parsing token")
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, nil
	}
	return nil, errors.New("invalid token")
}
