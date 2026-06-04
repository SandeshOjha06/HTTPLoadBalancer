package main

import (
	"sync/atomic"
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
