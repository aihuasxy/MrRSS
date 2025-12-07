package feed

import (
	"log"
	"time"
)

// Progress tracks the progress of feed fetching operations
type Progress struct {
	Total     int  `json:"total"`
	Current   int  `json:"current"`
	IsRunning bool `json:"is_running"`
}

// GetProgress returns the current progress of the feed fetching operation
func (f *Fetcher) GetProgress() Progress {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.progress
}

// waitForProgressComplete waits for any running operation to complete with a timeout.
// Returns true if the wait was successful, false if timeout occurred.
func (f *Fetcher) waitForProgressComplete(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for f.GetProgress().IsRunning {
		if time.Now().After(deadline) {
			log.Println("Timeout waiting for previous operation to complete")
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	return true
}
