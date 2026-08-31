package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secretKey []byte
	issuer    string
	expired   time.Duration
}
type Claims struct{
	jwt.RegisteredClaims
}
func NewJWTAthenticator(sKey []byte, iss string, exp time.Duration) JWTAuthenticator {
	return JWTAuthenticator{
		secretKey: sKey,
		issuer:    iss,
		expired:   exp,
	}
}
func (j *JWTAuthenticator) GenerateToken(userId string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userId,
		Issuer:    j.issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expired)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func (j *JWTAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
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
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
