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
}