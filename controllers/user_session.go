package controllers

import (
	"fmt"
	"time"

	"vvvorld/controllers/db"
	"vvvorld/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func encryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func LoginUser(c *gin.Context) {
	buffer := make([]byte, 2048)
	num, _ := c.Request.Body.Read(buffer)
	bodyRow := buffer[0:num]
	var body userLogin
	json.Unmarshal(bodyRow, &body)
	// fmt.Printf("user login: %+v \n", body)

	var search model.User
	result := db.Mdb.Where("email = ?", body.Email).First(&search)
	if result.Error != nil {
		fmt.Println("不存在该用户或者尚未注册")
		jsonRes := jsonResponse{
			-1,
			statusLoginNoUser,
			"no user error",
		}
		c.JSON(200, jsonRes)
		return
	}
	if comparePassword(search.PasswordHash, body.Password) != nil {
		fmt.Println("密码不正确")
		jsonRes := jsonResponse{
			-1,
			statusLoginIncorrectPass,
			"incorrect password",
		}
		c.JSON(200, jsonRes)
		return
	}
	_uuid, _ := uuid.NewRandom()
	uuid := _uuid.String()
	s := sessions.Default(c)
	s.Set(cookieLoginId, uuid)
	s.Set(cookieUserEmail, body.Email)
	sessionSaveErr := s.Save()
	if sessionSaveErr != nil {
		fmt.Println("session save error")
		jsonRes := jsonResponse{
			-1,
			statusLoginSessionSaveErr,
			"save session error",
		}
		c.JSON(200, jsonRes)
		return
	}
	// 测试用设置过期时间为1分钟
	// rdbSetErr := rdb.Set(c, uuid, "ok", 1*time.Minute).Err()
	rdbSetErr := db.Rdb.Set(c, uuid, "ok", 12*time.Hour).Err()
	if rdbSetErr != nil {
		fmt.Println("redis set session error")
		jsonRes := jsonResponse{
			-1,
			statusLoginRdbSetErr,
			"redis set error",
		}
		c.JSON(200, jsonRes)
		return
	}
	c.Redirect(303, "/")
	return
}

func SignupUser(c *gin.Context) {
	buffer := make([]byte, 2048)
	num, _ := c.Request.Body.Read(buffer)
	bodyRow := buffer[0:num]
	var body userSignup
	json.Unmarshal(bodyRow, &body)
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
		fmt.Println("未存在该email,继续执行操作")
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
	var newUser model.User
	newUser.UserName = body.UserName
	newUser.Email = body.Email
	newUser.PasswordHash = newPassHash
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

func LogoutUser(c *gin.Context) {
	s := sessions.Default(c)
	loginId := s.Get(cookieLoginId)
	if loginId == nil {
		fmt.Println("该用户已经退出登录")
		c.Redirect(303, "/login")
		return
	}
	db.Rdb.Del(c, loginId.(string))
	s.Delete(cookieLoginId)
	s.Delete(cookieUserEmail)
	s.Save()
	c.Redirect(303, "/login")
	return
}

func CheckLogin(c *gin.Context) (uint, error) {
	// return 0 -> not yet login
	// return 1 -> login
	s := sessions.Default(c)
	loginId := s.Get(cookieLoginId)
	if loginId == nil {
		fmt.Println("---请先登录---")
		return 0, nil
	}
	_, rdbGetErr := db.Rdb.Get(c, loginId.(string)).Result()
	if rdbGetErr == redis.Nil {
		fmt.Println("---session已过期---")
		return 0, nil
	} else if rdbGetErr != nil {
		fmt.Println("rbd 查找出错")
		return 0, rdbGetErr
	}
	return 1, nil
}

func CheckLOginMidllerware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		loginId := s.Get(cookieLoginId)
		if loginId == nil {
			c.AbortWithStatusJSON(403, gin.H{
				"msg": "please login ahead",
			})
			fmt.Println("---[api]请先登录---")
			return
		}
		_, rdbGetErr := db.Rdb.Get(c, loginId.(string)).Result()
		if rdbGetErr == redis.Nil {
			c.AbortWithStatusJSON(403, gin.H{
				"msg": "please login again",
			})
			fmt.Println("---[api]session已过期---")
			return
		} else if rdbGetErr != nil {
			c.AbortWithStatusJSON(403, gin.H{
				"msg": "please try again",
			})
			fmt.Println("---[api]rbd查找出错---")
			return
		}
		c.Next()
	}
}
