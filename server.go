package main

import (
	"fmt"
	"time"

	"vvvorld/controller"

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
	r.GET("/ws/:roomid", func(c *gin.Context) {
		roomid := c.Param("roomid")
		controller.WsController(c, roomid)
	})
	r.POST("/new_project", func(c *gin.Context) {
		buffer := make([]byte, 2048)
		num, _ := c.Request.Body.Read(buffer)
		fmt.Println(string(buffer[0:num]))
		c.JSON(200, gin.H{
			"msg": "create new project successfully",
		})
	})

	r.Run(":8080")
}
