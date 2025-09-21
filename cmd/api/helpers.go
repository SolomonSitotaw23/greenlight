package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/solomonsitotaw23/greenlight/internal/validator"
)

// an envelope type

type envelope map[string]any

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
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//Encode the data to json returning error if there was one.
	js, err := json.MarshalIndent(data, "", "\t")
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

func (app *application) readJson(w http.ResponseWriter, r *http.Request, dst any) error {
	// limit the size of request body to `1,048,576(1MB)` to protect from DOS
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_146)

	// to disallow requests out of our dst (this will throw an error if a user sends additional fields )
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body to the target destination
	err := dec.Decode(dst)
	if err != nil {
		//if there is an error during decoding
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var MaxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it
		// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		// for empty request
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty ")
		// if the json contains field that can not be mapped to the target dst
		case strings.HasPrefix(err.Error(), "json:unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s ", fieldName)
		// if the request body exceeds our size limit
		case errors.As(err, &MaxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", MaxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}

	}
	// check if the request body has additional JSON value
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must be only contain single JSON value")
	}
	return nil
}

// the readString() returns string value from the query string
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

// readCSV() function reads a string value from the query string and then splits it
// into a slice on the comma character.if no match found it returns the default value
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

// The readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// run a func as a go routine and prevent panic

func (app *application) background(fn func()) {
	// launch a background go routine
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		// Executing the the function
		fn()
	}()
}
