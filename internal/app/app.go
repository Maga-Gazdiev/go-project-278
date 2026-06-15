package app

import (
	"context"
	"fmt"
	"net/http"

	"project-3/internal/config"
	linkhandler "project-3/internal/handler/link"
	linkrepository "project-3/internal/repository/link"
	linkservice "project-3/internal/service/link"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	linkRepository := linkrepository.New(pool)
	linkService := linkservice.NewService(linkRepository, cfg.BaseURL)
	linkHandler := linkhandler.New(linkService)

	router := gin.Default()
	router.Use(corsMiddleware())
	linkhandler.RegisterRoutes(router, linkHandler)

	return router.Run(":" + cfg.Port)
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Range, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Range")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
