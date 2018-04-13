package main

import (
	"testing"
)

func TestGetToken(t *testing.T) {

	u := User{Email: "new@mail.com"}

	token := u.GetToken()
	if token != "0758993f5a28960962c1a521ceb928d259fe9c24c8b437be9ecd3e9458a4d4eb" {
		t.Errorf("Get Token return wrong result, expected: 0758993f5a28960962c1a521ceb928d259fe9c24c8b437be9ecd3e9458a4d4eb, got: %s", token)
	}
}
