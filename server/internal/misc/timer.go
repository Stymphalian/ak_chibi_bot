package misc

import (
	"log"
	"time"
)

// Helper function for running a callback function at a set interval
// Call the returned function to Close the callback.
func StartTimer(name string, interval time.Duration, callback func()) func() {
	ticker := time.NewTicker(interval)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				log.Println("Stopping timer", name)
				return
			case <-ticker.C:
				callback()
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(done)
	}
}
