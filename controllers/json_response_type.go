package controllers

type jsonResStatus uint

type jsonResponse struct {
	Code   int           `json:"code"` // 0 -> 成功, -1 -> 失败
	Status jsonResStatus `json:"status"`
	Msg    string        `json:"msg"`
}

// 2000番 -> 成功
//
//	4000番 -> 请求不正确
//
// 4100番 -> login相关, 4200番 -> signup相关
//
//	5000番 -> 服务端出错
//
// 5100番 -> login相关, 5200番 -> signup相关, 5300番 -> 登录后操作相关
const (
	statusLoginNoUser          jsonResStatus = 4101
	statusLoginIncorrectPass   jsonResStatus = 4102
	statusSignupExistingUser   jsonResStatus = 4201
	statusLoginSessionSaveErr  jsonResStatus = 5101
	statusLoginRdbSetErr       jsonResStatus = 5102
	statusSignupEncryptPassErr jsonResStatus = 5201
	statusSignupDbCreateErr    jsonResStatus = 5202
	statusGetUserErr           jsonResStatus = 5301
	statusJsonErr              jsonResStatus = 5302
	statusNewProjectErr        jsonResStatus = 5303
	statusGetProjectsErr       jsonResStatus = 5304
)

const (
	cookieLoginId   string = "loginId"
	cookieUserEmail string = "userEmail"
)
