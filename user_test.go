package main

import (
	"regexp"
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

func TestEmailRegex(t *testing.T) {
	reg := "(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)"

	tests := []struct {
		email string
		res   bool
	}{
		{
			"wrong",
			false,
		}, {
			"d@d.d",
			true,
		}, {
			"t@t",
			false,
		}, {
			"v@v.fl",
			true,
		}, {
			"vlad@webdeve.pro",
			true,
		}, {
			"adgasg@asdgasgadgasdg_asdgasg",
			false,
		},
	}

	for _, test := range tests {
		matched, _ := regexp.MatchString(reg, test.email)
		if matched != test.res {
			t.Errorf("%s expected: %t got %t", test.email, test.res, matched)
		}
	}
}
