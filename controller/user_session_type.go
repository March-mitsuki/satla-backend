package controller

type userLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userSignup struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 2000番 -> 成功

// 4000番 -> 请求不正确
// 4100番 -> login相关, 4200番 -> signup相关, 4300 -> session相关

// 5000番 -> 服务端出错
// 5100番 -> login相关, 5200番 -> signup相关, 5300 -> session相关
type responseStatus uint

const (
	statusLoginNoUser          responseStatus = 4101
	statusLoginIncorrectPass   responseStatus = 4102
	statusSignupExistingUser   responseStatus = 4201
	statusLoginSessionSaveErr  responseStatus = 5101
	statusLoginRdbSetErr       responseStatus = 5102
	statusSignupEncryptPassErr responseStatus = 5201
	statusSignupDbCreateErr    responseStatus = 5202
)
