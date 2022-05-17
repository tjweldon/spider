package util

import (
	"log"
	"net/http"
)

// MustGet is a wrapper around http.Get that exits the program if there is an
// error connecting to the http server.
func MustGet(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return resp
}

// GetOrNil returns a response pointer if the request was successful, nil
// otherwise.
func GetOrNil(url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}

	return resp
}
