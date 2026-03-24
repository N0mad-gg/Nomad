package channel

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c *gin.Context) {
	serverID := c.Param("server_id")
	userID := c.GetString("user_id")

	var exists bool
	h.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM server_members WHERE server_id = $1 AND user_id = $2)`,
		serverID, userID,
	).Scan(&exists)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT id, name, position FROM channels WHERE server_id = $1 ORDER BY position`,
		serverID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list channels"})
		return
	}
	defer rows.Close()

	type Channel struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Position int    `json:"position"`
	}

	channels := []Channel{}
	for rows.Next() {
		var ch Channel
		rows.Scan(&ch.ID, &ch.Name, &ch.Position)
		channels = append(channels, ch)
	}

	c.JSON(http.StatusOK, channels)
}

func (h *Handler) Create(c *gin.Context) {
	serverID := c.Param("server_id")
	userID := c.GetString("user_id")

	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role string
	err := h.db.QueryRow(context.Background(),
		`SELECT role FROM server_members WHERE server_id = $1 AND user_id = $2`,
		serverID, userID,
	).Scan(&role)
	if err != nil || (role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var channelID string
	h.db.QueryRow(context.Background(),
		`INSERT INTO channels (server_id, name) VALUES ($1, $2) RETURNING id`,
		serverID, req.Name,
	).Scan(&channelID)

	c.JSON(http.StatusCreated, gin.H{"id": channelID, "name": req.Name})
}

func (h *Handler) Delete(c *gin.Context) {
	serverID := c.Param("server_id")
	channelID := c.Param("channel_id")
	userID := c.GetString("user_id")

	var role string
	err := h.db.QueryRow(context.Background(),
		`SELECT role FROM server_members WHERE server_id = $1 AND user_id = $2`,
		serverID, userID,
	).Scan(&role)
	if err != nil || (role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	h.db.Exec(context.Background(),
		`DELETE FROM channels WHERE id = $1 AND server_id = $2`,
		channelID, serverID,
	)

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
