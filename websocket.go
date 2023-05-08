package torque

import (
	"context"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

// wsResponseWriter is a wrapper around the io.WriteCloser that
// implements the http.ResponseWriter interface.
type wsResponseWriter struct {
	wr io.WriteCloser
}

var _ http.ResponseWriter = (*wsResponseWriter)(nil)

func (w *wsResponseWriter) Header() http.Header {
	return map[string][]string{}
}

func (w *wsResponseWriter) Write(d []byte) (int, error) {
	return w.wr.Write(d)
}

func (*wsResponseWriter) WriteHeader(statusCode int) {}

// WebSocketParserFunc is a function that parses a websocket message
// and converts it into a *http.Request that can be handled by a route module.
type WebSocketParserFunc func(context.Context, string, int, []byte) (*http.Request, error)

func createWebsocketHandler(rm interface {
	Loader
	Renderer
}, parserFn WebSocketParserFunc, opts ...RouteOption) http.HandlerFunc {
	up := websocket.Upgrader{
		HandshakeTimeout: time.Second * 10,
	}
	rh := createRouteHandler(rm, opts...)

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := up.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "failed to upgrade connection to websocket", http.StatusInternalServerError)
			return
		}

		defer func() {
			err := ws.Close()
			if err != nil {
				log.Printf("[WebSocket] failed to close connection: %v\n", err)
			}
		}()

		for {
			mt, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("[WebSocket] failed to read message: %v\n", err)
				break
			}

			req, err := parserFn(r.Context(), r.URL.Path, mt, msg)
			if err != nil {
				log.Printf("[WebSocket] failed to parse message: %v\n", err)
				break
			}
			req.URL = r.URL
			req.RemoteAddr = r.RemoteAddr

			wr, err := ws.NextWriter(mt)
			if err != nil {
				log.Printf("[WebSocket] failed to get next writer: %v\n", err)
				break
			}

			rh.ServeHTTP(&wsResponseWriter{wr}, req)

			// TODO: if ServeHTTP panics, we should close the connection
			err = wr.Close()
			if err != nil {
				log.Printf("[WebSocket] failed to close writer: %v\n", err)
				break
			}
		}
	}
}
