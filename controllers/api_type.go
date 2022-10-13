package controllers

type responseUserInfo struct {
	Id       uint   `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
}

type createNewProjectBody struct {
	ID          uint   `json:"id"`
	ProjectName string `json:"project_name"`
	Description string `json:"description"`
	PointMan    string `json:"point_man"`
	CreatedBy   string `json:"created_by"`
}
