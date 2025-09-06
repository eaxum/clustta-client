package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GenerateToken() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const tokenLength = 64

	var token strings.Builder
	token.Grow(tokenLength)

	for i := 0; i < tokenLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err)
		}
		token.WriteByte(chars[n.Int64()])
	}

	return token.String()
}

func ValidateEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) (bool, string) {
	// Check password length
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	// Check for at least one uppercase letter
	var uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	if !uppercaseRegex.MatchString(password) {
		return false, "Password must contain at least one uppercase letter"
	}

	// Check for at least one lowercase letter
	var lowercaseRegex = regexp.MustCompile(`[a-z]`)
	if !lowercaseRegex.MatchString(password) {
		return false, "Password must contain at least one lowercase letter"
	}

	// Check for at least one digit
	var digitRegex = regexp.MustCompile(`[0-9]`)
	if !digitRegex.MatchString(password) {
		return false, "Password must contain at least one digit"
	}

	// Check for at least one special character
	var specialCharRegex = regexp.MustCompile(`[!@#~$%^&*()_+\-=\[\]{};':"\\|,.<>/?]+`)
	if !specialCharRegex.MatchString(password) {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
func CheckHashPassword(storedPassword string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
