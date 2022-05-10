package util

import (
	"log"
	"net/http"
)

func MustGet(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return resp
}
