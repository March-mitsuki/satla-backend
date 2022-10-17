package controllers

import (
	"fmt"
	"vvvorld/controllers/db"
	"vvvorld/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CreateNewUser(c *gin.Context) {
	buffer := make([]byte, 2048)
	num, _ := c.Request.Body.Read(buffer)
	bodyRow := buffer[0:num]
	var body userSignup
	unmarshalErr := json.Unmarshal(bodyRow, &body)
	if unmarshalErr != nil {
		fmt.Println("json解析出错")
		jsonRes := jsonResponse{
			-1,
			statusJsonErr,
			"json unmarshal error",
		}
		c.JSON(200, jsonRes)
		return
	}
	fmt.Printf("sign up body: %+v \n", body)

	var searchEmail model.User
	var searchName model.User
	searchEmailResult := db.Mdb.Where("email = ?", body.Email).First(&searchEmail)
	searchNameResult := db.Mdb.Where("user_name = ?", body.UserName).First(&searchName)
	if searchEmailResult.Error == nil && searchEmail.Email == body.Email {
		fmt.Println("该email已注册,请直接登录")
		jsonRes := jsonResponse{
			-1,
			statusSignupExistingUser,
			"existing user, please login",
		}
		c.JSON(200, jsonRes)
		return
	} else if searchEmailResult.Error != nil {
		fmt.Println("未存在该email,继续执行操作")
	}
	if searchNameResult.Error == nil && searchName.UserName == body.UserName {
		fmt.Println("该user name已注册,请直接登录")
		jsonRes := jsonResponse{
			-1,
			statusSignupExistingUser,
			"existing user, please login",
		}
		c.JSON(200, jsonRes)
		return
	} else if searchNameResult.Error != nil {
		fmt.Println("未存在该user name,继续执行操作")
	}

	newPassHash, encryptErr := encryptPassword(body.Password)
	if encryptErr != nil {
		fmt.Println("hash化密码失败")
		jsonRes := jsonResponse{
			-1,
			statusSignupEncryptPassErr,
			"encrypt password failed, please retry",
		}
		c.JSON(200, jsonRes)
		return
	}
	newUser := model.User{
		UserName:     body.UserName,
		Email:        body.Email,
		PasswordHash: newPassHash,
		Permission:   &body.Permission,
	}
	createResult := db.Mdb.Create(&newUser)
	if createResult.Error != nil {
		fmt.Println("创建用户失败")
		jsonRes := jsonResponse{
			-1,
			statusSignupDbCreateErr,
			"create new user failed, please retry",
		}
		c.JSON(200, jsonRes)
		return
	}
	c.Redirect(303, "/login")
	return
}

func CreateNewProject(c *gin.Context) {
	buffer := make([]byte, 2048)
	num, _ := c.Request.Body.Read(buffer)
	var body createNewProjectBody
	json.Unmarshal(buffer[0:num], &body)
	// fmt.Printf("\n unmarshal body: %+v \n", body)
	insertData := model.Project{
		ProjectName: body.ProjectName,
		Description: body.Description,
		PointMan:    body.PointMan,
		CreatedBy:   body.CreatedBy,
	}
	result := db.Mdb.Create(&insertData)
	if result.Error != nil {
		jsonRes := jsonResponse{
			-1,
			statusNewProjectErr,
			"db create err",
		}
		c.JSON(200, jsonRes)
		return
	}
	jsonRes := jsonResponse{
		0,
		2000,
		"create new project successfully",
	}
	c.JSON(200, jsonRes)
	return
}

func CheckAdminMiddlewareAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		email := s.Get(cookieUserEmail)
		var operatedUser model.User
		searchResult := db.Mdb.Where("email = ?", email).First(&operatedUser)
		if searchResult.Error != nil {
			c.AbortWithStatusJSON(403, gin.H{
				"msg": "can not found user",
			})
			fmt.Println("---[admin]查找user出错---")
			return
		}
		if *operatedUser.Permission != 2 {
			c.AbortWithStatusJSON(403, gin.H{
				"msg": "permission denied",
			})
			fmt.Println("---[admin]非管理员操作---")
			return
		}
		c.Next()
	}
}

func CheckAdminMiddlewarePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		email := s.Get(cookieUserEmail)
		var operatedUser model.User
		searchResult := db.Mdb.Where("email = ?", email).First(&operatedUser)
		if searchResult.Error != nil {
			c.Redirect(303, "/login")
			c.Abort()
			fmt.Println("---[admin-page]查找user出错---")
			return
		}
		if *operatedUser.Permission != 2 {
			c.Redirect(303, "/404")
			c.Abort()
			fmt.Println("---[admin-page]非管理员操作---")
			return
		}
		c.Next()
	}
}
