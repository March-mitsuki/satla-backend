package main

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"vvvorld/controllers"

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
	go controllers.WsHub.Run()
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
		num, err := controllers.CheckLogin(c)
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
		type testMsg struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		res := testMsg{
			Code: 0,
			Msg:  "server is running",
		}
		c.JSON(200, res)
	})
	r.GET("/ws/:roomid", func(c *gin.Context) {
		roomid := c.Param("roomid")
		// num, err := controllers.CheckLogin(c)
		// if err != nil {
		// 	fmt.Printf("check login err: %v \n", err)
		// 	return
		// }
		// if num == 0 {
		// 	return
		// }
		controllers.WsController(c, roomid)
		return
	})

	api := r.Group("/api")
	api.Use(controllers.CheckLOginMidllerware())
	api.POST("/new_project", func(c *gin.Context) {
		buffer := make([]byte, 2048)
		num, _ := c.Request.Body.Read(buffer)
		fmt.Println(string(buffer[0:num]))
		c.JSON(200, gin.H{
			"msg": "create new project successfully",
		})
	})
	api.GET("/crrent_userinfo", controllers.GetCurrentUserInfo)

	session := r.Group("/seesion")
	session.POST("/login", controllers.LoginUser)
	session.POST("/signup", controllers.SignupUser)
	session.DELETE("/logout", controllers.LogoutUser)

	r.NoRoute(func(c *gin.Context) {
		num, err := controllers.CheckLogin(c)
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
