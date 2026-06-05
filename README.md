# lb — HTTP Load Balancer

A production-grade HTTP load balancer built from scratch in Go, implementing round-robin scheduling, passive health checking, and graceful shutdown.

## What it does

- **Round-robin load balancing** — distributes requests evenly across backends using atomic operations
- **Passive health checking** — probes backends every 10 seconds, marks unreachable backends as DOWN
- **Automatic failover** — skips dead backends, routes only to healthy ones
- **Status endpoint** — `/status` returns JSON health report of all backends
- **CLI flags** — configurable port and backend list at runtime
- **Graceful shutdown** — catches SIGINT/SIGTERM, waits for in-flight requests before exiting

## Architecture

```
Client
  │
  ▼
[ ServeMux ]
  ├── /status ──────────────────→ statusHandler (JSON health report)
  └── /        ──────────────→ LoadBalancer.ServeHTTP
                                    │
                                    ▼
                              NextBackend()
                              (atomic round-robin)
                                    │
                          ┌─────────┼─────────┐
                          ▼         ▼         ▼
                      Backend 1  Backend 2  Backend 3
                      (alive)    (dead)     (alive)

Background:
  StartHealthCheck() ── ticker every 10s ──→ checkBackend() → setAlive()
```

## Key implementation details

**Thread-safe alive status** — each backend uses `sync.RWMutex`. Multiple goroutines can read `alive` concurrently (health checker + request handlers), but only one writes. `RLock` for reads, `Lock` for writes.

**Atomic round-robin** — uses `atomic.AddUint64` on a shared counter instead of a mutex. Avoids lock contention on the hot path. Index calculated as:

```go
idx := (next + uint64(i)) % uint64(total)
```

The `+ uint64(i)` walks forward from the starting point, so if backend N is dead it tries N+1, N+2, etc. without restarting the counter.

**Reverse proxy** — uses `httputil.NewSingleHostReverseProxy` from the standard library. Handles connection reuse, header forwarding, and response streaming automatically.

**Graceful shutdown sequence:**
1. Catch `SIGINT`/`SIGTERM` via `os.Signal` channel
2. Call `server.Shutdown(ctx)` with 30s timeout
3. HTTP server stops accepting new connections
4. Waits for in-flight requests to complete
5. Exits cleanly

**Health checker** — makes an HTTP GET to each backend with a 2s timeout. Marks alive if response is 2xx. Runs in a background goroutine with `time.NewTicker`.

## Build and run

```bash
git clone https://github.com/sandesh-ojha/lb
cd lb
go build -o lb .

# start with backends
./lb --port 8000 --backends http://localhost:8081,http://localhost:8082,http://localhost:8083
```

## Usage

```bash
# check backend health
curl http://localhost:8000/status

# output:
[
  {"url":"http://localhost:8081","alive":true},
  {"url":"http://localhost:8082","alive":false},
  {"url":"http://localhost:8083","alive":true}
]

# send requests (round-robins across alive backends)
curl http://localhost:8000
```

## Testing failover

```bash
# terminal 1 — start backends
python3 -m http.server 8081
python3 -m http.server 8082
python3 -m http.server 8083

# terminal 2 — start load balancer
./lb --port 8000 --backends http://localhost:8081,http://localhost:8082,http://localhost:8083

# terminal 3 — kill a backend and watch failover
kill $(lsof -t -i:8082)
curl http://localhost:8000  # still works, routes to 8081 and 8083
```

## Project structure

```
├── backend.go       # Backend struct, thread-safe alive methods
├── loadbalancer.go  # LoadBalancer, NextBackend, ServeHTTP, statusHandler
├── healthcheck.go   # checkBackend, StartHealthCheck
├── main.go          # CLI flags, wiring, graceful shutdown
└── go.mod
```

## What I learned

- Go concurrency primitives — goroutines, channels, `sync.RWMutex`, `sync/atomic`
- How reverse proxies work at the HTTP level
- Atomic operations for lock-free counters on hot paths
- Graceful shutdown with `context` and `os.Signal`
- Structuring a Go project across multiple files in one package

## References

- [net/http/httputil — Go standard library](https://pkg.go.dev/net/http/httputil)
- [sync/atomic — Go standard library](https://pkg.go.dev/sync/atomic)
- [How nginx load balancing works](https://nginx.org/en/docs/http/load_balancing.html)