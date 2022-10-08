package controllers

type responseUserInfo struct {
	Id       int    `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

type createNewProjectBody struct {
	ID          int64  `json:"id"`
	ProjectName string `json:"project_name"`
	Description string `json:"description"`
	PointMan    string `json:"point_man"`
	CreatedBy   string `json:"created_by"`
}
