package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(pass string) (string, error) {
	bytePass, err := bcrypt.GenerateFromPassword([]byte(pass), 12)
	if err != nil {
		return "", err
	}
	return string(bytePass), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}