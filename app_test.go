package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx"
)

type FakeStorage struct {
	name string
	data User
	body string
	code int
}

func (s FakeStorage) GetUserByEmail(u *User) error {
	if u.Email == "exist@user.com" {
		u.ID = 1
		return nil
	}

	if u.Email == "new@user.com" {
		return pgx.ErrNoRows
	}

	return fmt.Errorf("email do not match anything, please verify email address")
}

func (s FakeStorage) CreateUser(u *User) error {

	if u.Email == "exist@user.com" {
		return fmt.Errorf("user with such email address already exists")
	}

	if u.Email == "new@user.com" {
		u.ID = 1
		return nil
	}

	return fmt.Errorf("email do not match anything, please verify email address")
}

func SetUp(t *testing.T) *App {
	t.Parallel()
	a, _ := NewApp(&FakeStorage{})
	return &a
}

func executeRequest(a *App, req *http.Request) *httptest.ResponseRecorder {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "example.com")
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected int, response *httptest.ResponseRecorder, req *http.Request) {

	if expected != response.Code {
		i, _ := req.GetBody()
		b, err := ioutil.ReadAll(i)
		if err != nil {
			t.Errorf("Expected response code %d. Got %d\n, body read error: %v", expected, response.Code, err)
		}

		var msg interface{}
		err = json.Unmarshal(b, &msg)
		if err != nil {
			t.Errorf("Expected response code %d. Got %d\n, json unmarshal error: %v", expected, response.Code, err)
		}

		t.Errorf("Expected response code %d. Got %d\n, body: %+v", expected, response.Code, msg)
	}
}

func TestRegister(t *testing.T) {

	a := SetUp(t)
	tests := []FakeStorage{
		FakeStorage{
			name: "Email exists, password empty",
			data: User{
				Email:    "exist@user.com",
				Password: "",
			},
			body: `{"email":["email address already exists, do you want to reset password?"],"password":["cannot be empty"]}`,
			code: 400,
		},
		FakeStorage{
			name: "Email empty, password empty",
			data: User{
				Email:    "",
				Password: "",
			},
			body: `{"email":["cannot be empty"],"password":["cannot be empty"]}`,
			code: 400,
		},
		FakeStorage{
			name: "Email empty, password empty",
			data: User{
				Email:    "new@user.com",
				Password: "",
			},
			body: `{"password":["cannot be empty"]}`,
			code: 400,
		},
		FakeStorage{
			name: "Success requests",
			data: User{
				Email:    "new@user.com",
				Password: "123123",
			},
			body: `{"token":"65794a68624763694f694a49557a49314e694973496e523563434936496b705856434a392e65794a6c654841694f6a45314d4441774c434a7063334d694f694a755a58644164584e6c6369356a6232306966512e423064684e5450495f6d724138317a47346e7a7a357a32546a444541383348573461526e695a5679733155"}`,
			code: 201,
		},
	}

	for _, test := range tests {
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(test.data)

		req, _ := http.NewRequest("POST", "/register", b)
		response := executeRequest(a, req)

		checkResponseCode(t, test.code, response, req)

		if body := response.Body.String(); body != test.body {
			t.Errorf("%s, Expected %s but got '%s'", test.name, test.body, body)
			return
		}
	}
}

func TestSignUpOptions(t *testing.T) {
	a := SetUp(t)

	req, _ := http.NewRequest("OPTIONS", "/register", nil)
	response := executeRequest(a, req)
	checkResponseCode(t, 200, response, req)
}

func TestCORSHeaders(t *testing.T) {
	a := SetUp(t)

	req, _ := http.NewRequest("POST", "/register", nil)
	response := executeRequest(a, req)
	header := response.Header()

	tests := []struct {
		field string
		value string
	}{
		{
			field: "Access-Control-Allow-Origin",
			value: "example.com",
		}, {
			field: "Access-Control-Allow-Methods",
			value: "POST, GET, OPTIONS, PUT, DELETE",
		}, {
			field: "Access-Control-Allow-Headers",
			value: "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-REAL",
		}, {
			field: "Content-Type",
			value: "application/json",
		},
	}

	for _, test := range tests {
		if header.Get(test.field) != test.value {
			t.Errorf("Expected %s to be equal %v but got: %s",
				test.field,
				test.value,
				header.Get(test.field),
			)
		}
	}
}

func TestLogin(t *testing.T) {

	a := SetUp(t)
	tests := []FakeStorage{
		FakeStorage{
			name: "Empty email and password",
			data: User{
				Email:    "",
				Password: "",
			},
			body: `{"email":["email cannot be empty"],"password":["password cannot be empty"]}`,
			code: 400,
		},
		FakeStorage{
			name: "Wrong credentials",
			data: User{
				Email:    "new@user.com",
				Password: "bad password",
			},
			body: `{"__error__":["email or password do not match"]}`,
			code: 400,
		},
		FakeStorage{
			name: "Success requests",
			data: User{
				Email:    "exist@user.com",
				Password: "123123",
			},
			body: `{"token":"65794a68624763694f694a49557a49314e694973496e523563434936496b705856434a392e65794a6c654841694f6a45314d4441774c434a7063334d694f694a6c65476c7a64454231633256794c6d4e7662534a392e794f6a6a384b3171486559456d52505f7855526b483267384855483455746576436d694d5a555732314177"}`,
			code: 200,
		},
	}

	for _, test := range tests {
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(test.data)

		req, _ := http.NewRequest("POST", "/login", b)
		response := executeRequest(a, req)

		checkResponseCode(t, test.code, response, req)

		if body := response.Body.String(); body != test.body {
			t.Errorf("%s, Expected '%s' but got '%s'", test.name, test.body, body)
			return
		}
	}
}

func TestLoginOptions(t *testing.T) {
	a := SetUp(t)

	req, _ := http.NewRequest("OPTIONS", "/login", nil)
	response := executeRequest(a, req)
	checkResponseCode(t, 200, response, req)
}
