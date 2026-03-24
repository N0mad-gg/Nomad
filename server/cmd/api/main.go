package main

import (
	"log"

	"github.com/N0mad-gg/Nomad/server/config"
	"github.com/N0mad-gg/Nomad/server/internal/auth"
	"github.com/N0mad-gg/Nomad/server/internal/channel"
	"github.com/N0mad-gg/Nomad/server/internal/db"
	"github.com/N0mad-gg/Nomad/server/internal/message"
	nomadserver "github.com/N0mad-gg/Nomad/server/internal/server"
	"github.com/N0mad-gg/Nomad/server/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	pg := db.NewPostgres(cfg.DatabaseURL)
	defer pg.Close()

	_ = db.NewRedis(cfg.RedisURL)

	wtServer := &webtransport.Server{
		H3: &http3.Server{Addr: ":" + cfg.Port},
	}
	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, pg, wtServer)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://tauri.localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authHandler := auth.NewHandler(pg, cfg.JWTSecret)
	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		protected := api.Group("/")
		protected.Use(auth.Middleware(cfg.JWTSecret))
		{
			protected.GET("/me", func(c *gin.Context) {
				c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
			})

			serverHandler := nomadserver.NewHandler(pg)
			protected.POST("/servers", serverHandler.Create)
			protected.GET("/servers", serverHandler.List)
			protected.POST("/servers/join/:invite_code", serverHandler.Join)
			protected.DELETE("/servers/:server_id", serverHandler.Delete)

			channelHandler := channel.NewHandler(pg)
			protected.GET("/servers/:server_id/channels", channelHandler.List)
			protected.POST("/servers/:server_id/channels", channelHandler.Create)
			protected.DELETE("/servers/:server_id/channels/:channel_id", channelHandler.Delete)

			msgHandler := message.NewHandler(pg)
			protected.GET("/servers/:server_id/channels/:channel_id/messages", msgHandler.List)
			protected.POST("/servers/:server_id/channels/:channel_id/messages", msgHandler.Send)
			protected.DELETE("/servers/:server_id/channels/:channel_id/messages/:message_id", msgHandler.Delete)

			protected.GET("/gateway", wsHandler.Connect)
		}
	}

	log.Printf("Nomad server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
