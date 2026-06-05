package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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

    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
