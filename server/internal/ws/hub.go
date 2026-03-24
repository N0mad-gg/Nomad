package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/quic-go/webtransport-go"
)

type Event struct {
	Type      string `json:"type"`
	ChannelID string `json:"channel_id"`
	Payload   any    `json:"payload"`
}

type Client struct {
	userID    string
	serverIDs map[string]bool
	stream    *webtransport.Stream
	send      chan Event
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // userID -> Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(userID string, serverIDs []string, stream *webtransport.Stream) *Client {
	ids := make(map[string]bool)
	for _, id := range serverIDs {
		ids[id] = true
	}
	client := &Client{
		userID:    userID,
		serverIDs: ids,
		stream:    stream,
		send:      make(chan Event, 64),
	}
	h.mu.Lock()
	h.clients[userID] = client
	h.mu.Unlock()
	return client
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
}

func (h *Hub) BroadcastToServer(serverID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.serverIDs[serverID] {
			select {
			case client.send <- event:
			default:
			}
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-c.send:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if _, err := c.stream.Write(data); err != nil {
				log.Printf("write error for user %s: %v", c.userID, err)
				return
			}
		}
	}
}
