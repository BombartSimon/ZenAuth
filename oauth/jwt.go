package oauth

import (
	"time"
	"zenauth/config"

	"github.com/golang-jwt/jwt"
)

func GenerateAccessToken(subject string, scope string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   subject,
		"aud":   "zenauth",
		"scope": scope,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App.JWTSecret))
}

func ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.App.JWTSecret), nil
	})
}
