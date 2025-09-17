package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
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

// Global and Ip rate limiter
func (app *application) rateLimit(next http.Handler) http.Handler {

	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)

		// Lock the mutex to prevent this code from being executed concurrently.
		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = rate.NewLimiter(2, 4)
		}
		// Call the Allow() method on the rate limiter for the current IP address. If
		// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
		// response
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
