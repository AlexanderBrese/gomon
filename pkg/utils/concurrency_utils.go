package utils

import "sync"

func WithLock(mu *sync.Mutex, f func()) {
	mu.Lock()
	f()
	mu.Unlock()
}

func WithLockAndError(mu *sync.Mutex, f func() error) error {
	mu.Lock()
	if err := f(); err != nil {
		mu.Unlock()
		return err
	}
	mu.Unlock()
	return nil
}
