package torque

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sync"
)

var (
	ErrClosedByClient = errors.New("text/event-stream connection closed by client")
	ErrClosedByServer = errors.New("text/event-stream connection closed by server")
)

// WithEventStream creates a route that can support a text/event-stream connection. After a connection is
// established, the route will block until the client closes the connection or the server closes the connection.
//
// To send data to the client, send a struct that implements json.Marshaler to the channel.
//
// TODO(tyler): Refactor this to support more than just json.Marshaler -- I'd like to stream template fragments.
func WithEventStream(path string, ch chan json.Marshaler, cl chan error) RouteComponent {
	return func(r chi.Router) {
		r.HandleFunc(path, func(wr http.ResponseWriter, req *http.Request) {
			wg := sync.WaitGroup{}
			ctx := req.Context()

			flusher, ok := wr.(http.Flusher)
			if !ok {
				http.Error(wr, "client does not support text/event-stream", http.StatusInternalServerError)
				return
			}

			wr.Header().Set("Content-Type", "text/event-stream")
			wr.Header().Set("Cache-Control", "no-cache")
			wr.Header().Set("Connection", "keep-alive")
			wr.WriteHeader(200)

			flusher.Flush()

			wg.Add(1)
			go func() {
				defer wg.Done()

				<-ctx.Done()
				cl <- ErrClosedByClient
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					eventData, ok := <-ch
					if !ok {
						break
					}

					data, err := eventData.MarshalJSON()
					if err != nil {
						cl <- ErrClosedByServer

						break
					}

					_, err = fmt.Fprintf(wr, "data: %s\n\n", data)
					if err != nil {
						cl <- ErrClosedByServer

						break
					}

					flusher.Flush()
				}
			}()

			wg.Wait()
		})
	}
}
