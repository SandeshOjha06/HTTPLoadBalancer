package main

import (
	"log"
	"net/http"
	"time"
)

// checkBackend - makes an HTTP GET to the backend URL
// returns true if it responds with 2xx, false otherwise
func checkBackend(b *Backend) bool {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(b.url.String())
	if err != nil {
	return false
	}

	//close the response body
	defer resp.Body.Close()

	if resp.StatusCode >=200 && resp.StatusCode <= 299 {
	return true
	}

	return  false
}

// StartHealthCheck - runs forever in a goroutine
// every 10 seconds, checks all backends
func (lb *LoadBalancer) StartHealthCheck() {
	// create a ticker
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
	
		for _, b := range lb.backends {
			
			isAlive := checkBackend(b)

			b.setAlive(isAlive)

			if !isAlive {
				log.Printf("ALERT: Backend %s is DOWN", b.url.String())
			} else {
				log.Printf("SUCCESS: Backend %s is RUNNING", b.url.String())
			}
		}
	}
}
