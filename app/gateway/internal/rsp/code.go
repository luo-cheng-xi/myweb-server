package rsp

type statusCode = int

const (
	// 请求成功
	OK = 0
	// 用户注册相关
	EmailRegistered = 40001
	NameRegistered  = 40002
	// 用户登录相关
	PasswordWrong           = 40003
	UserNotExists           = 40004
	VerificationCodeInvalid = 40005
	EmptyParam              = 40006
	Internal                = 500
)
