package broadcast

import (
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebsocketBroadcaster contains information on the websocket that we are communicating with
type WebsocketBroadcaster struct {
	m          sync.Mutex
	websockets []messageWriteCloser
	rchan      <-chan []byte
	qchan      chan struct{}
}

type messageWriteCloser interface {
	WriteMessage(int, []byte) error
	Close() error
}

// New creates a new WebsocketBroadcaster
func New(rchan <-chan []byte) *WebsocketBroadcaster {
	return &WebsocketBroadcaster{
		rchan: rchan,
		qchan: make(chan struct{}),
	}
}

// Add adds the websocket information to the WebsocketBroadcaser stuct
func (w *WebsocketBroadcaster) Add(ws messageWriteCloser) {
	w.m.Lock()
	defer w.m.Unlock()

	w.websockets = append(w.websockets, ws)
}

// Broadcast sends data to the websocket
func (w *WebsocketBroadcaster) Broadcast() {
	go func() {
		for {
			select {
			case msg := <-w.rchan:
				w.m.Lock()
				for i := len(w.websockets) - 1; i >= 0; i-- {
					ws := w.websockets[i]
					err := ws.WriteMessage(websocket.TextMessage, msg)
					if err != nil {
						log.WithFields(log.Fields{
							"Error": err,
						}).Debug("Error writing to websocket")
						w.websockets = w.websockets[:i+copy(w.websockets[i:], w.websockets[i+1:])]
						ws.Close()
						w.m.Unlock()
					}
				}
				w.m.Unlock()
			case <-w.qchan:
				w.m.Lock()
				for _, ws := range w.websockets {
					ws.Close()
				}
				w.m.Unlock()
				w.qchan <- struct{}{}
				return
			}
		}
	}()
}

// Stop handles closing the websocket.
func (w *WebsocketBroadcaster) Stop() chan struct{} {
	w.qchan <- struct{}{}
	return w.qchan
}
