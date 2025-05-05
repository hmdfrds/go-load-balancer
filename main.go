package main

import (
	"net/url"
	"sync"
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

func main() {

}
