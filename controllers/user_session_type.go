package controllers

type userLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userSignup struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
