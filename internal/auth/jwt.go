package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTInterface interface {
	GenerateToken(userId string) (string, error)
	ValidateToken(tokenString string) (string, error)
}

type JWTService struct {
	secretKey []byte
	issuer    string
	expired   time.Duration
}
func NewJWTService(sKey []byte, iss string, exp time.Duration) *JWTService {
	return &JWTService{
		secretKey: sKey,
		issuer:    iss,
		expired:   exp,
	}
}

func (j *JWTService) GenerateToken(userId string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userId,
		Issuer:    j.issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expired)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {

			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}

			return j.secretKey, nil
		},

		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	if claims.Subject == "" {
		return "", errors.New("invalid token subject")
	}

	return claims.Subject, nil
}
