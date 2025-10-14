package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordCost defines the bcrypt cost factor (12 for strong security)
const PasswordCost = 12

// HashPassword hashes a password using bcrypt with cost factor 12
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
