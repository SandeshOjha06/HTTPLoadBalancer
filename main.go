package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8000, "port to listen on")
	backends := flag.String("backends", "", "list of backends")

	flag.Parse()

	if *backends == "" {
		log.Fatal("--backends list is reqd")
	}

	backendList := strings.Split(*backends, ",")

	lb := &LoadBalancer{}

	for _, serverStr := range backendList {
		serverUrl, err := url.Parse(serverStr)
		if err != nil {
			log.Fatalf("Failed to parse URL %s: %v", serverStr, err)
		}
		backend := &Backend{
			url:   serverUrl,
			alive: true,
		}
		lb.backends = append(lb.backends, backend)
	}

	go lb.StartHealthCheck()

	fmt.Printf("Load balancer started on http://localhost:%d\n", *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/status", lb.statusHandler)
	mux.Handle("/", lb)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
	}
    //  Start the server in a background goroutine so it doesn't block the rest of main()
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // blocks unitil ctrl+c

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}