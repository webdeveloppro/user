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
			body: `{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld0B1c2VyLmNvbSIsImZpcnN0X25hbWUiOiIiLCJsYXN0X25hbWUiOiIifQ.knY8cZ7UbOYaQHwaqJKQejiQIieyROibKZihLY4iQNk"}`,
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
			body: `{"email":["cannot be empty"],"password":["cannot be empty"]}`,
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
			body: `{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImV4aXN0QHVzZXIuY29tIiwiZmlyc3RfbmFtZSI6IiIsImxhc3RfbmFtZSI6IiJ9.qDDYj2lOcom9KNF06GxOygc9QJ-7W47GPLnkCHl5fDI"}`,
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

func TestProfile(t *testing.T) {

	type ProfileStruct struct {
		name  string
		token string
		body  string
		code  int
	}

	u := User{}
	u.Email = "new@email.com"
	token, _ := u.generateToken()

	a := SetUp(t)
	tests := []ProfileStruct{
		ProfileStruct{
			name:  "Without token",
			token: "",
			body:  `{"error":"Authorization"}`,
			code:  401,
		},
		ProfileStruct{
			name:  "Wrong token",
			token: "123123",
			body:  `{"error":"token contains an invalid number of segments"}`,
			code:  401,
		},
		ProfileStruct{
			name:  "Success requests",
			token: token,
			body:  `{"email":"` + u.Email + `","first_name":"","last_name":""}`,
			code:  200,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", "/profile", nil)

		if test.token != "" {
			req.Header.Set("Authorization", test.token)
		}

		response := executeRequest(a, req)

		// checkResponseCode(t, test.code, response, req)

		if body := response.Body.String(); body != test.body {
			t.Errorf("%s, Expected '%s' but got '%s'", test.name, test.body, body)
			return
		}
	}
}

func TestProfileOptions(t *testing.T) {
	a := SetUp(t)

	req, _ := http.NewRequest("OPTIONS", "/profile", nil)
	response := executeRequest(a, req)
	checkResponseCode(t, 200, response, req)
}
