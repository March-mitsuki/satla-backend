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
