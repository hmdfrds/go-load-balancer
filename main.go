package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL   *url.URL
	Alive bool
}

type ServerPool struct {
	backends []*Backend
	current  uint64
	mux      sync.RWMutex
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.backends = append(s.backends, backend)
	log.Printf("Added backend: %s", backend.URL)
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			if b.Alive != alive {
				b.Alive = alive
				status := "DOWN"
				if alive {
					status = "UP"
				}
				log.Printf("Backend %s status changed to [%s]", backendUrl.String(), status)
			}
			return
		}
	}
}

func (s *ServerPool) GetNextPeer() *Backend {
	s.mux.RLock()
	defer s.mux.RUnlock()

	numBackends := len(s.backends)
	if numBackends == 0 {
		log.Println("No backends available in the pool.")
		return nil
	}

	nextIdx := atomic.AddUint64(&s.current, 1) - 1
	startIdx := nextIdx % uint64(numBackends)

	for i := 0; i < numBackends; i++ {
		currentIndex := (startIdx + uint64(i)) % uint64(numBackends)
		backend := s.backends[currentIndex]

		if backend.Alive {
			return backend
		}
	}

	log.Printf("No healthy backends available.")
	return nil
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Backend %s unreachable, error: %s", u.Host, err)
		return false
	}
	_ = conn.Close()
	return true
}

func healthCheck(pool *ServerPool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Starting health check pass...")
		pool.mux.RLock()
		backendsToCheck := make([]*Backend, len(pool.backends))
		copy(backendsToCheck, pool.backends)
		pool.mux.RUnlock()

		var wg sync.WaitGroup
		for _, backend := range backendsToCheck {
			wg.Add(1)
			go func(b *Backend) {
				defer wg.Done()
				alive := isBackendAlive(b.URL)
				pool.MarkBackendStatus(b.URL, alive)
			}(backend)
		}
		wg.Wait()
		log.Println("Health check pass complete.")
	}
}

func lb(pool *ServerPool, w http.ResponseWriter, r *http.Request) {
	peer := pool.GetNextPeer()
	if peer == nil {
		log.Printf("Request %s %s: No healthy peers available", r.Method, r.URL.Path)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Request %s %s: Routing to backend %s", r.Method, r.URL.Path, peer.URL)
	proxy := httputil.NewSingleHostReverseProxy(peer.URL)

	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = peer.URL.Scheme
		r.URL.Host = peer.URL.Host

		r.Host = peer.URL.Host

		r.Header.Set("X-Real-IP", getIP(r.RemoteAddr))
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

		fwdFor := r.Header.Get("X-Forwarded-For")
		clientIP := getIP(r.RemoteAddr)
		if fwdFor != "" {
			fwdFor = fwdFor + ", " + clientIP
		} else {
			fwdFor = clientIP
		}
		r.Header.Set("X-Forwarded-For", fwdFor)
		log.Printf("Proxying request for %s to %s%s", r.RemoteAddr, r.URL.Host, r.URL.Path)

	}

	proxy.ErrorHandler = func(ew http.ResponseWriter, er *http.Request, err error) {
		log.Printf("Proxy error: Backend %s - %v", peer.URL, err)
		http.Error(ew, "Bad Gateway", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

func getIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func main() {
	backendUrls := []string{
		"http://localhost:9001",
		"http://localhost:9002",
	}

	listenAddr := ":8080" // load balancer port

	pool := ServerPool{}

	log.Println("Configuration backends:")
	for _, urlStr := range backendUrls {
		backendUrl, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Error parsing backend URL '%s': '%v'", urlStr, err)
		}
		backend := &Backend{URL: backendUrl}
		pool.AddBackend(backend)
	}

	log.Println("Starting background health checker...")
	go healthCheck(&pool)

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		lb(&pool, w, r)
	}

	log.Printf("Load Balancer server starting on %s", listenAddr)
	server := &http.Server{
		Addr:    listenAddr,
		Handler: http.HandlerFunc(httpHandler),
	}

	log.Fatal(server.ListenAndServe())

}
