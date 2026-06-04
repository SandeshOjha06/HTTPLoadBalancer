package main

import (
    "log"
    "net/http"
    "net/url"
    "fmt"
)

func main() {
    servers := []string{
        "http://localhost:8081",
        "http://localhost:8082",
        "http://localhost:8083",
    }

    lb := &LoadBalancer{}

    for _, serverStr := range servers {
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

    fmt.Println("Load balancer started on :8000")

    server := http.Server{
        Addr:    ":8000",
        Handler: lb,
    }

    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}