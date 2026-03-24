package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,min=2,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	inviteCode := generateInviteCode()

	var serverID string
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO servers (name, owner_id, invite_code) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, userID, inviteCode,
	).Scan(&serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create server"})
		return
	}

	h.db.Exec(context.Background(),
		`INSERT INTO server_members (server_id, user_id, role) VALUES ($1, $2, 'owner')`,
		serverID, userID,
	)

	h.db.Exec(context.Background(),
		`INSERT INTO channels (server_id, name, position) VALUES ($1, 'general', 0)`,
		serverID,
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":          serverID,
		"name":        req.Name,
		"invite_code": inviteCode,
	})
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(context.Background(),
		`SELECT s.id, s.name, s.icon_url, s.invite_code
		 FROM servers s
		 JOIN server_members sm ON sm.server_id = s.id
		 WHERE sm.user_id = $1
		 ORDER BY sm.joined_at`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list servers"})
		return
	}
	defer rows.Close()

	type Server struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		IconURL    *string `json:"icon_url"`
		InviteCode string  `json:"invite_code"`
	}

	servers := []Server{}
	for rows.Next() {
		var s Server
		rows.Scan(&s.ID, &s.Name, &s.IconURL, &s.InviteCode)
		servers = append(servers, s)
	}

	c.JSON(http.StatusOK, servers)
}

func (h *Handler) Join(c *gin.Context) {
	inviteCode := c.Param("invite_code")
	userID := c.GetString("user_id")

	var serverID string
	err := h.db.QueryRow(context.Background(),
		`SELECT id FROM servers WHERE invite_code = $1`,
		inviteCode,
	).Scan(&serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid invite code"})
		return
	}

	_, err = h.db.Exec(context.Background(),
		`INSERT INTO server_members (server_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		serverID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"server_id": serverID})
}

func (h *Handler) Delete(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.GetString("user_id")

	result, err := h.db.Exec(context.Background(),
		`DELETE FROM servers WHERE id = $1 AND owner_id = $2`,
		serverID, userID,
	)
	if err != nil || result.RowsAffected() == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func generateInviteCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
