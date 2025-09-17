package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

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

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// A background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enable {

			ip := realip.FromRequest(r)

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			if _, found := clients[ip]; !found {
				// Create and add a new client struct to the map if it doesn't already exist.
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}
			// update the lastseen time for the client
			clients[ip].lastSeen = time.Now()
			// Call the Allow() method on the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}
