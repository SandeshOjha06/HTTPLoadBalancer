package main

import (
	"net/url"
	"sync"
)

type Backend struct {
url *url.URL
alive bool
mux sync.RWMutex
}

func(b *Backend) setAlive(alive bool){
	// use Lock() because to lock writing to prevent race condn
b.mux.Lock()
b.alive = alive
defer b.mux.Unlock()
}


func(b *Backend) isAlive() bool {
	// to check isalive only neer to read, so using RLock
b.mux.RLock()
defer b.mux.RUnlock()
return b.alive
} 
