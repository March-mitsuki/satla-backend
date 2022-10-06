package controllers

type responseUserInfo struct {
	Id       int    `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}
