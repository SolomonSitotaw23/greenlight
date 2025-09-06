package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Retrieve the "id" URL parameter from the current request context
func (app *application) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// helper for sending responses this takes the destination http.ResponseWritter and the http status code
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	//Encode the data to json returning error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append new line to make it easier for terminal applications
	js = append(js, '\n')
	// loop through the header map and add each header to the http.ResponseWriter header map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add content type application/json then write the status code and json response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
