package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
	Here is test using real database base connection
	We want to do end 2 end tests to make use we have right database schema
	We are not testing login here - just http responses
*/

func MakeRequest(url, reqType string, expectedCode int, data interface{}) (map[string]interface{}, error) {

	var jsonResp map[string]interface{}
	b := new(bytes.Buffer)

	if data != nil {
		json.NewEncoder(b).Encode(data)
	}

	req, err := http.NewRequest(
		reqType,
		url,
		b,
	)

	if err != nil {
		return jsonResp, fmt.Errorf("Cannot open %s %s: %v", reqType, url, err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	res, err := client.Do(req)

	if err != nil {
		return jsonResp, fmt.Errorf("Cannot make request %s %s: %v", reqType, url, err)
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return jsonResp, fmt.Errorf("Response body read error: %v", err)
	}

	err = json.Unmarshal(bodyBytes, &jsonResp)
	if err != nil {
		return jsonResp, fmt.Errorf("Json unmarshal error: %v, body: %s", err, bodyBytes)
	}

	if res.StatusCode != expectedCode {
		return jsonResp, fmt.Errorf("Request %s %s, return wrong status, expected: %d, got: %d", reqType, url, expectedCode, res.StatusCode)
	}

	return jsonResp, nil
}

var validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3QzQHRlc3QuY29tIiwiZmlyc3RfbmFtZSI6IiIsImxhc3RfbmFtZSI6IiJ9.NlQ5xAWNQm9nrB5etqFMgQeTBKU8f_5LHREnAH3aOcE"

func TestRegister(t *testing.T) {

	serverUrl := fmt.Sprintf("http://%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	url := serverUrl + "/register"
	reqType := "POST"
	code := 201

	data := map[string]string{
		"email":    "test3@test.com",
		"password": "123123123",
	}

	resData, err := MakeRequest(url, reqType, code, data)
	if err != nil {
		t.Errorf("err: %v\n data: %+v", err, resData)
		return
	}

	if resData["token"] != validToken {
		t.Errorf("Response does not match, expect: %s, got: %v", validToken, resData)
	}

	code = 400

	resData, err = MakeRequest(url, reqType, code, data)
	if err != nil {
		t.Errorf("err: %v\n data: %+v", err, resData)
		return
	}

	if resData["email"] == "" {
		t.Errorf("Response does not match, expect email not empty, %+v", resData)
		return
	}
}

func TestLogin(t *testing.T) {

	// ToDo
	// Load fixtures

	serverUrl := fmt.Sprintf("http://%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	url := serverUrl + "/login"
	reqType := "POST"
	code := 200

	data := map[string]string{
		"email":    "test3@test.com",
		"password": "123123123",
	}

	resData, err := MakeRequest(url, reqType, code, data)
	if err != nil {
		t.Errorf("err: %v\n data: %+v", err, resData)
		return
	}

	if resData["token"] != validToken {
		t.Errorf("Response does not match, expect: %s, got: %v", validToken, resData)
		return
	}

	data["password"] = "123123123123213"
	code = 400

	resData, err = MakeRequest(url, reqType, code, data)
	if err != nil {
		t.Errorf("err: %v\n data: %+v", err, resData)
		return
	}

	if resData["__error__"] == "" {
		t.Errorf("Response does not match, expect have an error, got: %+v", resData)
	}
}
