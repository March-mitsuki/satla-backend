package controllers

type responseUserInfo struct {
	Id         uint   `json:"id"`
	UserName   string `json:"user_name"`
	Email      string `json:"email"`
	Permission uint   `json:"permission"`
}

type reqChangePassBody struct {
	Id       uint   `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	OldPass  string `json:"old_password"`
	NewPass  string `json:"new_password"`
}
