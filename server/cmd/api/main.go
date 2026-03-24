package main

import (
	"log"

	"github.com/N0mad-gg/Nomad/server/config"
	"github.com/N0mad-gg/Nomad/server/internal/auth"
	"github.com/N0mad-gg/Nomad/server/internal/channel"
	"github.com/N0mad-gg/Nomad/server/internal/db"
	nomadserver "github.com/N0mad-gg/Nomad/server/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	pg := db.NewPostgres(cfg.DatabaseURL)
	defer pg.Close()

	_ = db.NewRedis(cfg.RedisURL)

	r := gin.Default()

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
			protected.DELETE("/servers/:id", serverHandler.Delete)

			channelHandler := channel.NewHandler(pg)
			protected.GET("/servers/:server_id/channels", channelHandler.List)
			protected.POST("/servers/:server_id/channels", channelHandler.Create)
			protected.DELETE("/servers/:server_id/channels/:channel_id", channelHandler.Delete)
		}
	}

	log.Printf("Nomad server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
