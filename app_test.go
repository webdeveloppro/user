package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
			body: "null",
			code: 204,
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
			body: `{"token":"d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86"}`,
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

func TestUploadFile(t *testing.T) {
	a := SetUp(t)

	tests := []struct {
		name        string
		token       string
		fileName    string
		fileContent []byte
		code        int
		body        []byte
		location    string
	}{
		{
			name:        "Success transaction",
			token:       "d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86",
			fileName:    "readme.txt",
			fileContent: []byte("readme before doing anything"),
			code:        201,
			body:        []byte("[]"),
			location:    "/content/readme.txt",
		}, {
			name:        "Wrong token string",
			token:       "",
			fileName:    "readme.txt",
			fileContent: []byte(""),
			code:        403,
			body:        []byte(`{"__error__":"Token is required"}`),
			location:    "",
		},
	}

	for _, test := range tests {

		req, _ := http.NewRequest("PUT", "/files/"+test.fileName, bytes.NewReader(test.fileContent))
		req.Header.Set("Authorization", test.token)
		response := executeRequest(a, req)

		if test.code != response.Code {
			t.Errorf("%s, Response code does not match expect %d got %d", test.name, test.code, response.Code)
		} else if string(test.body) != response.Body.String() {
			t.Errorf("%s, Response body does not match expect %s got %s", test.name, test.body, response.Body.String())
		} else if test.location != response.Header().Get("Location") {
			t.Errorf("%s, Response location does not match expect %s got %s", test.name, test.location, response.Header().Get("Location"))
		}
	}
}

func TestGetFile(t *testing.T) {
	a := SetUp(t)

	createReadMeFile(t)
	tests := []struct {
		name        string
		token       string
		fileName    string
		code        int
		body        []byte
		contentType string
	}{
		{
			name:        "Successful request",
			token:       "d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86",
			fileName:    "readme.txt",
			code:        200,
			body:        []byte("read me before doing anything"),
			contentType: "application/octet-stream",
		}, {
			name:        "Wrong token string",
			token:       "",
			fileName:    "readme.txt",
			code:        403,
			body:        []byte(`{"__error__":"Token is required"}`),
			contentType: "application/json",
		}, {
			name:        "File does not exists",
			token:       "d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86",
			fileName:    "readme2.txt",
			code:        404,
			body:        []byte(`File not found`),
			contentType: "",
		},
	}

	for _, test := range tests {

		req, _ := http.NewRequest("GET", "/files/"+test.fileName, nil)
		req.Header.Set("Authorization", test.token)
		response := executeRequest(a, req)

		if test.code != response.Code {
			t.Errorf("%s, Response code does not match expect %d got %d", test.name, test.code, response.Code)
		} else if string(test.body) != response.Body.String() {
			t.Errorf("%s, Response body does not match expect %s got %s", test.name, test.body, response.Body.String())
		} else if test.contentType != response.Header().Get("Content-Type") {
			t.Errorf("%s, Response content type does not match expect %s got %s", test.name, test.contentType, response.Header().Get("Content-Type"))
		}
	}
}

func TestDeleteFile(t *testing.T) {
	a := SetUp(t)

	createReadMeFile(t)
	tests := []struct {
		name     string
		token    string
		fileName string
		code     int
		body     []byte
	}{
		{
			name:     "Successful request",
			token:    "d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86",
			fileName: "readme.txt",
			code:     204,
			body:     []byte(""),
		}, {
			name:     "File does not exists",
			token:    "d82303a16a459fd9cc9b5d2a01e5a3730812676839f877757991f63adb2a5d86",
			fileName: "readme2.txt",
			code:     404,
			body:     []byte(`File not found`),
		},
	}

	for _, test := range tests {

		req, _ := http.NewRequest("DELETE", "/files/"+test.fileName, nil)
		req.Header.Set("Authorization", test.token)
		response := executeRequest(a, req)

		if test.code != response.Code {
			t.Errorf("%s, Response code does not match expect %d got %d", test.name, test.code, response.Code)
		} else if string(test.body) != response.Body.String() {
			t.Errorf("%s, Response body does not match expect %s got %s", test.name, test.body, response.Body.String())
		}
	}
}

func createReadMeFile(t *testing.T) {
	_ = os.Remove("./content/readme.txt")
	buf := bytes.NewReader([]byte("read me before doing anything"))
	f, err := os.OpenFile("./content/readme.txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Errorf("Cannot open file, %v", err)
		return
	}
	defer f.Close()
	io.Copy(f, buf)
}
