package misc

import (
	"log"
	"time"
)

// Helper function for running a callback function at a set interval
// Call the returned function to Close the callback.
func StartTimer(name string, interval time.Duration, callback func()) func() {
	ticker := time.NewTicker(interval)
	stop := make(chan bool)
	done := make(chan bool)
	go func() {
		// GoRunCounter.Add(1)
		// defer GoRunCounter.Add(-1)

		for {
			select {
			case <-stop:
				log.Println("Stopping timer", name)
				done <- true
				return
			case <-ticker.C:
				callback()
			}
		}
	}()

	return func() {
		ticker.Stop()
		close(stop)
		<-done
	}
}
