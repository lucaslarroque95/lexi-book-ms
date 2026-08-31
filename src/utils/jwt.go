package utils

import (
	"crypto/rsa"
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Keys struct {
	publicKey *rsa.PublicKey
}

func (k *Keys) LoadPublicKey(path string) error {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return (err)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return (err)
	}

	k.publicKey = key
	return nil
}

func (k *Keys) VerifyToken(token string) error {

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return k.publicKey, nil
	})

	if err != nil {
		return errors.New("could not parse token")
	}

	if !parsedToken.Valid {
		return errors.New("invalid token")
	}

	return nil
}

func (k *Keys) ExtractClaims(token string) (string, []string, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
		return "", nil, errors.New("could not parse token")
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		return "", nil, errors.New("invalid userId claim")
	}

	roleClaim, _ := claims["role"].(string)
	var roles []string
	if roleClaim != "" {
		roles = strings.Split(roleClaim, ",")
	}

	return userId, roles, nil
}
