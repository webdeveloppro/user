package user

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
	stringToken, err := u.generateToken()
	if err != nil {
		return "", errors.Wrapf(err, "user: cannot generate jwt token")
	}

	return stringToken, nil
}

// generateToken will generate token and return byte array
func (u *User) generateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      u.Email,
		"first_name": "",
		"last_name":  "",
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(hmacSecret)
}

// InvalidToken will check if user have valid token
func (u *User) InvalidToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSecret, nil
	})

	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		u.Email = claims["email"].(string)
		return true, nil
	}

	// Since we don't have any instruction for the token i assume if token is not empty its valid
	return false, nil
}
