package utils

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// Batcher batches file watch events in a given interval.
type Batcher struct {
	*fsnotify.Watcher
	interval time.Duration
	done     chan struct{}

	Events chan []fsnotify.Event
	Errors chan []error
}

// NewBatcher creates and starts a Batcher with the given time interval.
func NewBatcher(interval time.Duration) (*Batcher, error) {
	watcher, err := fsnotify.NewWatcher()

	batcher := &Batcher{}
	batcher.Watcher = watcher
	batcher.interval = interval
	batcher.done = make(chan struct{}, 1)
	batcher.Events = make(chan []fsnotify.Event, 1)

	if err == nil {
		go batcher.run()
	}

	return batcher, err
}

func (b *Batcher) run() {
	tick := time.Tick(b.interval)
	evs := make([]fsnotify.Event, 0)
	errs := make([]error, 0)
OuterLoop:
	for {
		select {
		case ev := <-b.Watcher.Events:
			evs = append(evs, ev)
		case err := <-b.Watcher.Errors:
			errs = append(errs, err)
		case <-tick:
			if len(evs) != 0 {
				b.Events <- evs
				evs = make([]fsnotify.Event, 0)
			}
			if len(errs) != 0 {
				b.Errors <- errs
				errs = make([]error, 0)
			}

		case <-b.done:
			break OuterLoop
		}
	}
	close(b.done)
}

func (b *Batcher) Close() {
	b.done <- struct{}{}
	b.Watcher.Close()
}
