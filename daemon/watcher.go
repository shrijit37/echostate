package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceInterval = 500 * time.Millisecond

type Watcher struct {
	watcher *fsnotify.Watcher
	state   *StateManager
	stop    chan struct{}
}

func NewWatcher(state *StateManager, filename string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch the directory containing the file
	dir := filepath.Dir(state.mdPath)
	if err := w.Add(dir); err != nil {
		w.Close()
		return nil, err
	}

	watcher := &Watcher{
		watcher: w,
		state:   state,
		stop:    make(chan struct{}),
	}

	go watcher.run(filename)
	return watcher, nil
}

func (w *Watcher) run(filename string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C // Drain initial
	needsRefresh := false

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == filename {
				needsRefresh = true
				debounceTimer.Reset(debounceInterval)
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		case <-debounceTimer.C:
			if needsRefresh {
				if err := w.state.Refresh(); err != nil {
					log.Printf("refresh: %v", err)
				}
				needsRefresh = false
			}
		case <-ticker.C:
			// Periodic check for file existence / manual edits outside fsnotify
			if _, err := os.Stat(w.state.mdPath); err == nil && w.state.mdContent == "" {
				if err := w.state.Refresh(); err != nil {
					log.Printf("refresh: %v", err)
				}
			}
		case <-w.stop:
			return
		}
	}
}

func (w *Watcher) Close() {
	close(w.stop)
	w.watcher.Close()
}
