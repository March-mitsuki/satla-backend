package controllers

type responseUserInfo struct {
	Id       uint   `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}
