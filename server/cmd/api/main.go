package main

import (
	"log"

	"github.com/N0mad-gg/Nomad/server/config"
	"github.com/N0mad-gg/Nomad/server/internal/auth"
	"github.com/N0mad-gg/Nomad/server/internal/db"
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
		}
	}

	log.Printf("Nomad server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
