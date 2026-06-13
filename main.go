package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Gender   string `json:"gender"`
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return router
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	router := setupRouter()
	router.Run(":" + port)
}
