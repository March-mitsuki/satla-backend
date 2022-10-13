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
		Id:       user.ID,
		UserName: user.UserName,
		Email:    user.Email,
	}
	c.JSON(200, res)
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
