package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3131"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept-Encoding",
			"User-Agent",
		},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.StaticFile("/", "./dist")
	r.Static("/assets", "./dist/assets")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8080")
}
