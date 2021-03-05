package utils

import "sync"

func WithLock(mu *sync.Mutex, f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}
