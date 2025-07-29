package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters long")
	}

	if len(password) > 72 {
		return "", fmt.Errorf("password must be at most 72 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CheckPasswordMatch(password, hashedPassword string) bool {
	err := CheckPassword(password, hashedPassword)
	return err == nil
}
