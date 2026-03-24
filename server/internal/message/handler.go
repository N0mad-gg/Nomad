package message

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

type Message struct {
	ID        string  `json:"id"`
	ChannelID string  `json:"channel_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"created_at"`
	EditedAt  *string `json:"edited_at"`
}

func (h *Handler) isMember(serverID, userID string) bool {
	var exists bool
	h.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM server_members WHERE server_id = $1 AND user_id = $2)`,
		serverID, userID,
	).Scan(&exists)
	return exists
}

func (h *Handler) List(c *gin.Context) {
	serverID := c.Param("server_id")
	channelID := c.Param("channel_id")
	userID := c.GetString("user_id")

	if !h.isMember(serverID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT m.id, m.channel_id, m.user_id, u.username, m.content,
		        m.created_at::text, m.edited_at::text
		 FROM messages m
		 JOIN users u ON u.id = m.user_id
		 WHERE m.channel_id = $1
		 ORDER BY m.created_at DESC
		 LIMIT $2`,
		channelID, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		rows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.Username, &m.Content, &m.CreatedAt, &m.EditedAt)
		messages = append(messages, m)
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) Send(c *gin.Context) {
	serverID := c.Param("server_id")
	channelID := c.Param("channel_id")
	userID := c.GetString("user_id")

	if !h.isMember(serverID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=2000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var m Message
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO messages (channel_id, user_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, channel_id, user_id, content, created_at::text`,
		channelID, userID, req.Content,
	).Scan(&m.ID, &m.ChannelID, &m.UserID, &m.Content, &m.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	h.db.QueryRow(context.Background(),
		`SELECT username FROM users WHERE id = $1`, userID,
	).Scan(&m.Username)

	c.JSON(http.StatusCreated, m)
}

func (h *Handler) Delete(c *gin.Context) {
	serverID := c.Param("server_id")
	messageID := c.Param("message_id")
	userID := c.GetString("user_id")

	if !h.isMember(serverID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	result, err := h.db.Exec(context.Background(),
		`DELETE FROM messages WHERE id = $1 AND user_id = $2`,
		messageID, userID,
	)
	if err != nil || result.RowsAffected() == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
