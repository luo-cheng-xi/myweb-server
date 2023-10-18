package model

type UserVO struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}
type UserVoWithToken struct {
	UserVO
	Token string `json:"token"`
}
