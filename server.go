package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"vvvorld/controllers"
	"vvvorld/controllers/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadDotenv() error {
	// 读取VVVRD_ENV和GIN_MODE, 若都为空则设置为开发模式
	// 如果
	const (
		development string = "development"
		production  string = "production"
		test        string = "test"
	)
	env := os.Getenv("VVVRD_ENV")
	ginMode := os.Getenv("GIN_MODE")
	if env == "" && ginMode == "" {
		env = development
	} else if ginMode == "release" {
		env = production
	}
	switch env {
	case development:
		fmt.Println("[notice] dotenv is runnning on 'dev' mode")
		err := godotenv.Load(".env." + env + ".local")
		if err != nil {
			return err
		}
	case production:
		fmt.Println("[notice] dotenv is runnning on 'production' mode")
		err := godotenv.Load(".env." + env)
		if err != nil {
			return err
		}
	case test:
		fmt.Println("[notice] dotenv is runnning on 'test' mode")
		err := godotenv.Load(".env.local")
		if err != nil {
			return err
		}
	}
	err := godotenv.Load()
	if err != nil {
		fmt.Println("[notice] default .env file is undifine")
	}
	return nil
}

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
	dotenvErr := loadDotenv()
	if dotenvErr != nil {
		fmt.Printf("\n [error] dotenv load err: %v \n", dotenvErr)
		panic("--- dotenv loading err ---")
	}
	dbConnErr := db.ConnectionDB()
	if dbConnErr != nil {
		fmt.Printf("\n [error] db connection err: %v \n", dbConnErr)
		panic("--- db connection error ---")
	}

	go controllers.WsHub.Run()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{os.Getenv("CORS_ORIGIN")},
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
		controllers.WsController(c, roomid)
		return
	})

	api := r.Group("/api")
	if os.Getenv("GIN_MODE") == "release" {
		// 在release模式中api启用登录检测
		api.Use(controllers.CheckLOginMidllerware())
	}
	api.GET("/crrent_userinfo", controllers.GetCurrentUserInfo)
	api.GET("/all_projects", controllers.GetAllProjects)

	session := r.Group("/seesion")
	session.POST("/login", controllers.LoginUser)
	session.DELETE("/logout", controllers.LogoutUser)

	admin := r.Group("/admin")
	admin.POST("/new_user", controllers.CreateNewUser)
	admin.POST("/new_project", controllers.CreateNewProject)

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
