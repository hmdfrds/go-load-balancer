package main

import (
	"log"
	"net/url"
	"sync"
	"sync/atomic"
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

func main() {

}
