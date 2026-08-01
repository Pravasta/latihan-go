package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is set above bcrypt.DefaultCost (10) to raise the work factor
// for offline brute-force attempts against a leaked user store.
const bcryptCost = 14

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)

	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
