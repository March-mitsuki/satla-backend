package controllers

type userSignup struct {
	UserName        string `json:"user_name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	Permission      uint   `json:"permission"`
	OperatedByEmail string `json:"operated_by_email"`
}

type createNewProjectBody struct {
	ID              uint   `json:"id"`
	ProjectName     string `json:"project_name"`
	Description     string `json:"description"`
	PointMan        string `json:"point_man"`
	CreatedBy       string `json:"created_by"`
	OperatedByEmail string `json:"operated_by_email"`
}
