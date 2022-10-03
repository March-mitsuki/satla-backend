package main

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"vvvorld/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func directAccess(c *gin.Context) {
	_, file := path.Split(c.Request.RequestURI)
	ext := filepath.Ext(file)
	if file == "" || ext == "" {
		c.File("./dist" + "/index.html")
	} else {
		c.File("./dist" + c.Request.RequestURI)
	}
	return
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3131"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
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
	sessionStore := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("loginInfo", sessionStore))

	r.Static("/assets", "./dist/assets")
	r.GET("/", func(c *gin.Context) {
		num, err := controller.CheckLogin(c)
		if err != nil {
			c.Redirect(303, "/login")
			fmt.Printf("check login rdb err: %v \n", err)
			return
		}
		if num == 0 {
			c.Redirect(303, "/login")
			return
		}
		c.File("./dist")
		return
	})
	r.GET("/login", directAccess)
	r.GET("/signup", directAccess)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./public/favicon.ico")
		return
	})

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server is running!",
		})
	})
	r.GET("/ws/:roomid", func(c *gin.Context) {
		roomid := c.Param("roomid")
		controller.WsController(c, roomid)
	})

	api := r.Group("/api")
	api.POST("/new_project", func(c *gin.Context) {
		buffer := make([]byte, 2048)
		num, _ := c.Request.Body.Read(buffer)
		fmt.Println(string(buffer[0:num]))
		c.JSON(200, gin.H{
			"msg": "create new project successfully",
		})
	})
	api.POST("/login", controller.LoginUser)
	api.POST("/signup", controller.SignupUser)
	api.DELETE("/logout", controller.LogoutUser)

	r.NoRoute(func(c *gin.Context) {
		num, err := controller.CheckLogin(c)
		if err != nil {
			c.Redirect(303, "/login")
			fmt.Printf("check login rdb err: %v \n", err)
			return
		}
		if num == 0 {
			c.Redirect(303, "/login")
			return
		}
		directAccess(c)
		return
	})

	r.Run(":8080")
}
