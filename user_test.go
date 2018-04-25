package main

import (
	"testing"
)

func TestGetToken(t *testing.T) {

	u := User{Email: "new@mail.com"}

	expectedToken := "65794a68624763694f694a49557a49314e694973496e523563434936496b705856434a392e65794a6c654841694f6a45314d4441774c434a7063334d694f694a755a586441625746706243356a6232306966512e43545744696f7861515877315f5f4247712d495a64364d6a3338454f48496b4e726e343839763470442d77"
	token, err := u.GetToken()
	if err != nil {
		t.Errorf("Error happen: %v", err)
	}

	if token != expectedToken {
		t.Errorf("Get Token return wrong result, expected: %s, got: %s", expectedToken, token)
	}
}
