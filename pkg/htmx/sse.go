package htmx

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrNotSupported   = fmt.Errorf("text/event-stream not available to client")
	ErrClosedByClient = fmt.Errorf("text/event-stream connection closed by client")
	ErrClosedByServer = fmt.Errorf("text/event-stream connection closed by server")
)

// EventKey is a string that identifies an event sent to an eventsource.
type EventKey = string

// SourceFunc is a closure that provides a channel that can be used to
// send data to an eventsource client. Typically paired with a keyed event
// name in an EventSourceMap.
type SourceFunc func(chan string)

// EventSourceMap is a map of event names to event source functions.
type EventSourceMap map[EventKey]SourceFunc

// SSE creates a handler that is adapted to htmx's sse extension.
func SSE(wr http.ResponseWriter, req *http.Request, sources EventSourceMap) error {
	loo, ok := wr.(http.Flusher)
	if !ok {
		return ErrNotSupported
	}

	wr.Header().Set("Content-Type", "text/event-stream")
	wr.Header().Set("Cache-Control", "no-cache")
	wr.Header().Set("Connection", "keep-alive")
	wr.WriteHeader(http.StatusOK)

	loo.Flush()
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for key, fn := range sources {
		wg.Add(1)
		go func(mu *sync.Mutex, wg *sync.WaitGroup, key EventKey, fn SourceFunc) {
			defer wg.Done()

			ch := make(chan string)
			defer close(ch)

			go fn(ch)

			for {
				select {
				case <-req.Context().Done():
					return

				case data, ok := <-ch:
					if !ok {
						return
					}

					// lock the mutex to prevent concurrent writes to the
					// response writer before flushing
					mu.Lock()

					_, err := wr.Write([]byte("event: " + key + "\n"))
					if err != nil {
						return
					}

					// One must split the data by newline and write each line
					// with 'data:' prepended to it. Kind of silly but keeps the
					// response from being malformed.
					// https://discord.com/channels/725789699527933952/1166121680301731900
					lines := strings.Split(strings.TrimSpace(data), "\n")
					for _, line := range lines {
						_, err = wr.Write([]byte("data: " + line + "\n"))
						if err != nil {
							return
						}
					}

					// write the final newline to delineate the end of the message
					_, err = wr.Write([]byte("\n"))
					if err != nil {
						return
					}

					loo.Flush()
					mu.Unlock()
				}
			}
		}(&mu, &wg, key, fn)
	}
	wg.Wait()
	return nil
}
