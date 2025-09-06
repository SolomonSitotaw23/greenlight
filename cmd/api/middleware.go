package main

import (
	"fmt"
	"net/http"
)

// a middleware to recover panic by responding to the client
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// this defer func always runs as go unwinds the stack
			defer func() {
				// check if there has been a panic or not by using the builtin recover function
				if err := recover(); err != nil {
					// If there was a panic, set a "Connection: close" header on the
					// response
					w.Header().Set("Connection", "close")

					app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
				}
			}()
			next.ServeHTTP(w, r)
		})
}
