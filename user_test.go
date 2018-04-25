package main

import (
	"testing"
)

func TestGetToken(t *testing.T) {

	u := User{Email: "new@mail.com"}

	expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld0BtYWlsLmNvbSIsImZpcnN0X25hbWUiOiIiLCJsYXN0X25hbWUiOiIifQ.wv7XDm_1m4jS1MW5q_3sE8yxGmGw6Oo7fB0NQ9S6T0A"
	token, err := u.GetToken()
	if err != nil {
		t.Errorf("Error happen: %v", err)
	}

	if token != expectedToken {
		t.Errorf("Get Token return wrong result, expected: %s, got: %s", expectedToken, token)
	}
}
