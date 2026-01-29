package gateway

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// client represents one websocket connection.
type client struct {
	ws *websocket.Conn     // underlying socket
	tx chan []byte         // outbound queue
	id string              // just RemoteAddr for demo
}

// hub tracks all connected clients and broadcasts messages.
type hub struct {
	mu       sync.RWMutex
	clients  map[*client]bool
	register chan *client
}

// newHub starts an event loop that keeps the maps in sync.
func newHub() *hub {
	h := &hub{
		clients:  make(map[*client]bool),
		register: make(chan *client, 16),
	}
	go h.run()
	return h
}

func (h *hub) run() {
	for c := range h.register {
		h.mu.Lock()
		h.clients[c] = true
		h.mu.Unlock()
		log.Debug().Str("id", c.id).Msg("new websocket client")
	}
}

// broadcast encodes v as JSON and sends it to every client.
func (h *hub) broadcast(v interface{}) {
	data, _ := json.Marshal(v)

	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.tx <- data:
		default:
			// connection is stuck – drop it
			close(c.tx)
			delete(h.clients, c)
			_ = c.ws.Close()
		}
	}
	h.mu.RUnlock()
}