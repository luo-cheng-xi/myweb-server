package rsp

import "myweb/app/gateway/internal/model"

type BaseRsp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
	Data       any    `json:"data"`
}

type TokenRsp struct {
	BaseRsp
	Token string `json:"token"`
}
type UserAndTokenRsp struct {
	model.UserVO

	Token string `json:"token"`
}
