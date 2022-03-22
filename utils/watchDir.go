package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watches a directory recursively
// Caller must call .Close
func NewDirWatcher(path string) (*Batcher, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := Batcher{Watcher: watcher, interval: time.Millisecond * 100, done: make(chan struct{}), Events: make(chan []fsnotify.Event)}

	watchDir := func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return nil
		}
		if !fi.Mode().IsDir() {
			return nil
		}

		w.Watcher.Add(path)
		// since fsnotify can watch all the files in a directory, watchers only need
		// to be added to each nested directory

		return nil
	}
	go w.run()
	if false {

		err = w.Watcher.Add(path)
	} else {
		if err := filepath.Walk(path, watchDir); err != nil {
			return nil, err
		}

	}
	return &w, err
}

type watch struct {
	watcher *fsnotify.Watcher
}

// watchDir gets run as a walk func, searching for directories to add watchers to

// Batcher batches file watch events in a given interval.
type Batcher struct {
	Watcher  *fsnotify.Watcher
	interval time.Duration
	done     chan struct{}

	Events chan []fsnotify.Event // Events are returned on this channel
}

func (b *Batcher) run() {
	tick := time.Tick(b.interval)
	defer b.Watcher.Close()
	evs := make([]fsnotify.Event, 0)
OuterLoop:
	for {
		select {
		case ev, ok := <-b.Watcher.Events:
			if !ok {
				continue
			}
			evs = append(evs, ev)
		// case err, ok := <-b.Watcher.Errors:
		// 	if !ok {
		// 		continue
		// 	}
		case <-tick:
			if len(evs) == 0 {
				continue
			}
			b.Events <- evs
			evs = make([]fsnotify.Event, 0)
		case <-b.done:
			break OuterLoop
		}
	}
	close(b.done)
}
func (b *Batcher) Close() {
	b.done <- struct{}{}
}
