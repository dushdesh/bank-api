package util

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	// Return the hashed password as a string
	return string(hashedPassword), nil
}

// CheckPassword checks if the password is correct
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
