package utils

import "sync"

// WithLock runs the given operation in a blocking manner
func WithLock(mu *sync.Mutex, f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}

// WithLockAndError runs the given operation in a blocking manner and returns any errors
func WithLockAndError(mu *sync.Mutex, f func() error) error {
	mu.Lock()
	if err := f(); err != nil {
		mu.Unlock()
		return err
	}
	mu.Unlock()
	return nil
}
