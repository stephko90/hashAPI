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

func TestFindHashInDatabase(t *testing.T) {
	id := 123
	hash := "abcdf"
	db := strings.NewReader("123 abcdf")

	val := findHashInDatabase(db, id)

	// hit
	if val != hash {
		t.Fail()
	}

	id = 1000
	val = findHashInDatabase(db, id)

	// miss
	if val != "" {
		t.Fail()
	}
}

func TestGetLineCount(t *testing.T) {
	numLines := 3
	reader := strings.NewReader("test\ntest2\ntest3")

	lines := getLineCount(reader)
	if lines != numLines {
		t.Fail()
	}
}

func TestLoadTotalTime(t *testing.T) {
	time := "123"
	timeDb := strings.NewReader(time)

	loadTotalTime(timeDb)

	timeInt, _ := strconv.Atoi(time)
	if totalTime != int64(timeInt) {
		t.Fail()
	}
}