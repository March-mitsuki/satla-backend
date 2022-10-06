package controller

import (
	"vvvorld/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetCurrentUserInfo(c *gin.Context) {
	s := sessions.Default(c)
	email := s.Get(cookieUserEmail)
	var user model.User
	result := db.Where("email = ?", email).First(&user)
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
