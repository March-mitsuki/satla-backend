package controllers

import (
	"fmt"
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
		jsonRes := jsonResponse{
			-1,
			statusGetUserErr,
			"db get error",
		}
		c.JSON(200, jsonRes)
		return
	}
	res := responseUserInfo{
		Id:         user.ID,
		UserName:   user.UserName,
		Email:      user.Email,
		Permission: *user.Permission,
	}
	c.JSON(200, res)
}

func GetAllProjects(c *gin.Context) {
	var projects []model.Project
	result := db.Mdb.Find(&projects)
	if result.Error != nil {
		jsonRes := jsonResponse{
			-1,
			statusGetProjectsErr,
			"db find err",
		}
		c.JSON(200, jsonRes)
		return
	}
	c.JSON(200, projects)
	return
}

func ChangeUserPassword(c *gin.Context) {
	buffer := make([]byte, 2048)
	num, _ := c.Request.Body.Read(buffer)
	bodyRow := buffer[0:num]
	var body reqChangePassBody
	unmarshalErr := json.Unmarshal(bodyRow, &body)
	if unmarshalErr != nil {
		jsonRes := jsonResponse{
			-1,
			statusJsonErr,
			"json unmarshal error",
		}
		c.JSON(200, jsonRes)
		return
	}
	arg := db.ArgChangeUserPassword{
		ID:      body.Id,
		OldPass: body.OldPass,
		NewPass: body.NewPass,
	}
	err := db.ChangeUserPassword(arg)
	if err != nil {
		jsonRes := jsonResponse{
			-1,
			statusChangePassDbErr,
			fmt.Sprintln(err),
		}
		c.JSON(200, jsonRes)
		return
	}
	jsonRes := jsonResponse{
		0,
		statusSuccessful,
		"password change successfully",
	}
	c.JSON(200, jsonRes)
	return
}
