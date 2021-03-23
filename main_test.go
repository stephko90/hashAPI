package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestGetIdFromHashUrl(t *testing.T) {
	id := "123"
	url := "/path/to/route/"+id

	vars := strings.Split(url, "/")
	newId := getIdFromHashUrl(vars)
	newIdStr := strconv.Itoa(newId)
	if newIdStr != id {
		t.Fail()
	}

	badUrl := "/test"
	vars = strings.Split(badUrl, "/")
	newId = getIdFromHashUrl(vars)
	if newId != -1 {
		t.Fail()
	}
}

func TestConvertPass(t *testing.T) {
	pass := "test"

	converted := convertPass(pass)
	// logic here is that base64 strings always end with = or ==
	if !strings.Contains(converted, "=") {
		t.Fail()
	}
}

func TestValidatePassword(t *testing.T) {
	valid, err := validatePassword("")

	if valid || err == "" {
		t.Fail()
	}

	valid, err = validatePassword("12345678901")

	if valid || err == "" {
		t.Fail()
	}

	valid, err = validatePassword("o'reily")

	if valid || err == "" {
		t.Fail()
	}

	valid, err = validatePassword("test")

	if !valid || err != "" {
		t.Fail()
	}
}