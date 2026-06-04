package main

import (
	"sync/atomic"
	"net/http"
	"net/http/httputil"
)

type LoadBalancer struct {
backends []*Backend
current uint64
}

func (lb *LoadBalancer) NextBackend() *Backend {
    total := len(lb.backends)
    if total == 0 {
        return nil
    }
	// find out the next index
    next := atomic.AddUint64(&lb.current, 1)
    
	// look through the backends[]
    for i := 0; i < total; i++ {
		// take an index
        idx := (next + uint64(i)) % uint64(total)
        //grab the server
		backend := lb.backends[idx]
        if backend.isAlive() {
            return backend
        }
    }
    return nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.NextBackend()
	if peer == nil {
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
	return
	}

	proxy := httputil.NewSingleHostReverseProxy(peer.url)
    proxy.ServeHTTP(w, r)
}
