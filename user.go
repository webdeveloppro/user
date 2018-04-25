package main

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

var hmacSecret = []byte("588b3236da217f94682121eeeb2732b204a083c5b8a417fe3e58c7072ef81b6b")

// User information
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// GetToken will return X-Session token
func (u *User) GetToken() (string, error) {
	byteToken, err := u.generateToken()
	if err != nil {
		return "", errors.Wrapf(err, "user: cannot generate jwt token")
	}

	return fmt.Sprintf("%x", byteToken[:]), nil
}

// generateToken will generate token and return byte array
func (u *User) generateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: 15000,
		Issuer:    u.Email,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(hmacSecret)
}

// InvalidToken will check if user have valid token
func InvalidToken(token string) bool {
	// Since we don't have any instruction for the token i assume if token is not empty its valid
	return token == ""
}
