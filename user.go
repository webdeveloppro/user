package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// User information
type User struct {
	ID        int32     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// GetToken will return X-Session token
func (u *User) GetToken() string {
	byteToken := u.generateToken()
	return fmt.Sprintf("%x", byteToken[:])
}

// generateToken will generate token and return byte array
func (u *User) generateToken() [32]byte {
	return sha256.Sum256([]byte(u.Email))
}

// InvalidToken will check if user have valid token
func InvalidToken(token string) bool {
	// Since we don't have any instruction for the token i assume if token is not empty its valid
	return token == ""
}
