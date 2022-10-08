package controllers

import (
	"vvvorld/controllers/db"
	"vvvorld/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetCurrentUserInfo(c *gin.Context) {
	s := sessions.Default(c)
	email := s.Get(cookieUserEmail)
	var user model.User
	result := db.Mdb.Where("email = ?", email).First(&user)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"code":   -1,
			"status": statusGetUserErr,
			"msg":    "db get error",
		})
		return
	}
	res := responseUserInfo{
		Id:       int(user.ID),
		UserName: user.UserName,
		Email:    user.Email,
	}
	c.JSON(200, res)
}

func CreateNewProject(c *gin.Context) {
	buffer := make([]byte, 0)
	num, _ := c.Request.Body.Read(buffer)
	var body createNewProjectBody
	json.Unmarshal(buffer[0:num], &body)
	insertData := model.Project{
		ProjectName: body.ProjectName,
		Description: body.Description,
		PointMan:    body.PointMan,
		CreatedBy:   body.CreatedBy,
	}
	result := db.Mdb.Create(&insertData)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"code":   -1,
			"status": statusNewProjectErr,
			"msg":    "db create err",
		})
		return
	}
	c.JSON(200, gin.H{
		"code":   0,
		"status": 2000,
		"msg":    "create new project successfully",
	})
}
