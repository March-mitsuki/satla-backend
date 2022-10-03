package controller

import (
	"fmt"
	"time"

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
	result := db.Where("email = ?", body.Email).First(&search)
	if result.Error != nil {
		fmt.Println("不存在该用户或者尚未注册")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 4101,
			"msg":    "no user error",
		})
		return
	}
	if comparePassword(search.PasswordHash, body.Password) != nil {
		fmt.Println("密码不正确")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 4102,
			"msg":    "incorrect password",
		})
		return
	}
	_uuid, _ := uuid.NewRandom()
	uuid := _uuid.String()
	s := sessions.Default(c)
	s.Set("loginId", uuid)
	sessionSaveErr := s.Save()
	if sessionSaveErr != nil {
		fmt.Println("session save error")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 5101,
			"msg":    "save session error",
		})
		return
	}
	// 测试用设置过期时间为1分钟
	// rdbSetErr := rdb.Set(c, uuid, "ok", 1*time.Minute).Err()
	rdbSetErr := rdb.Set(c, uuid, "ok", 12*time.Hour).Err()
	if rdbSetErr != nil {
		fmt.Println("redis set session error")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 5102,
			"msg":    "redis set error",
		})
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

	var search model.User
	searchResult := db.Where("email = ?", body.Email).First(&search)
	if searchResult.Error == nil && search.Email == body.Email {
		fmt.Println("已存在该用户,请直接登录")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 4201,
			"msg":    "existing user, please login",
		})
		return
	} else if searchResult.Error != nil {
		fmt.Println("未存在该用户,继续执行操作")
	}
	newPassHash, encryptErr := encryptPassword(body.Password)
	if encryptErr != nil {
		fmt.Println("hash化密码失败")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 5201,
			"msg":    "encrypt password failed, please retry",
		})
		return
	}
	var newUser model.User
	newUser.UserName = body.UserName
	newUser.Email = body.Email
	newUser.PasswordHash = newPassHash
	createResult := db.Create(&newUser)
	if createResult.Error != nil {
		fmt.Println("创建用户失败")
		c.JSON(200, gin.H{
			"code":   -1,
			"status": 5202,
			"msg":    "create new user failed, please retry",
		})
		return
	}
	c.Redirect(303, "/login")
	return
}

func LogoutUser(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete("loginId")
	s.Save()
	c.Redirect(303, "/login")
	return
}

func CheckLogin(c *gin.Context) (uint, error) {
	// 0 -> redirect to login page
	// 1 -> go next
	s := sessions.Default(c)
	sInfo := s.Get("loginId")
	if sInfo == nil {
		fmt.Println("---请先登录---")
		return 0, nil
	}
	_, rdbGetErr := rdb.Get(c, sInfo.(string)).Result()
	if rdbGetErr == redis.Nil {
		fmt.Println("---session已过期---")
		return 0, nil
	} else if rdbGetErr != nil {
		fmt.Println("rbd 查找出错")
		return 0, rdbGetErr
	}
	c.File("./dist")
	return 1, nil
}