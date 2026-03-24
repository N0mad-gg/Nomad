package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quic-go/webtransport-go"
)

type Handler struct {
	hub    *Hub
	db     *pgxpool.Pool
	server *webtransport.Server
}

func NewHandler(hub *Hub, db *pgxpool.Pool, wtServer *webtransport.Server) *Handler {
	return &Handler{hub: hub, db: db, server: wtServer}
}

func (h *Handler) Connect(c *gin.Context) {
	userID := c.GetString("user_id")

	session, err := h.server.Upgrade(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebTransport upgrade failed"})
		return
	}

	ctx := session.Context()
	stream, err := session.AcceptStream(ctx)
	if err != nil {
		log.Printf("accept stream error: %v", err)
		return
	}

	serverIDs := h.getUserServerIDs(userID)
	client := h.hub.Register(userID, serverIDs, stream)
	defer h.hub.Unregister(userID)

	go client.WritePump(ctx)
	h.readPump(ctx, userID, stream)
}

func (h *Handler) readPump(ctx context.Context, userID string, stream *webtransport.Stream) {
	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			return
		}

		var event struct {
			Type      string `json:"type"`
			ChannelID string `json:"channel_id"`
			Content   string `json:"content"`
		}
		if err := json.Unmarshal(buf[:n], &event); err != nil {
			continue
		}

		if event.Type == "message" && event.Content != "" && event.ChannelID != "" {
			h.handleMessage(ctx, userID, event.ChannelID, event.Content)
		}
	}
}

func (h *Handler) handleMessage(ctx context.Context, userID, channelID, content string) {
	var serverID, msgID, username, createdAt string
	err := h.db.QueryRow(ctx,
		`SELECT c.server_id FROM channels c WHERE c.id = $1`, channelID,
	).Scan(&serverID)
	if err != nil {
		return
	}

	var exists bool
	h.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM server_members WHERE server_id = $1 AND user_id = $2)`,
		serverID, userID,
	).Scan(&exists)
	if !exists {
		return
	}

	err = h.db.QueryRow(ctx,
		`INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`,
		channelID, userID, content,
	).Scan(&msgID, &createdAt)
	if err != nil {
		return
	}

	h.db.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username)

	h.hub.BroadcastToServer(serverID, Event{
		Type:      "message",
		ChannelID: channelID,
		Payload: map[string]any{
			"id":         msgID,
			"user_id":    userID,
			"username":   username,
			"content":    content,
			"created_at": createdAt,
		},
	})
}

func (h *Handler) getUserServerIDs(userID string) []string {
	rows, err := h.db.Query(context.Background(),
		`SELECT server_id FROM server_members WHERE user_id = $1`, userID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids
}
