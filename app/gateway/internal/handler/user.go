package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"myweb/app/api/pb"
	"myweb/app/gateway/internal/model"
	"myweb/app/gateway/internal/rsp"
	"myweb/app/gateway/rpc"
	"net/http"
	"time"
)

const (
	UserClientKey = "userClient"
)

type UserHandler struct {
	logger   *zap.Logger
	userConn *rpc.UserConn
}

func NewUserHandler(logger *zap.Logger, userConn *rpc.UserConn) *UserHandler {
	return &UserHandler{
		logger:   logger,
		userConn: userConn,
	}
}

// Register 用户注册，发送注册邮件
func (u *UserHandler) Register(ctx *gin.Context) {
	// 读取传来的PostForm信息 | read the given post form information
	userName := ctx.PostForm("name")
	userPassword := ctx.PostForm("password")
	userEmail := ctx.PostForm("email")
	// 包装rpc请求信息
	registerReq := pb.RegisterReq{
		Name:     userName,
		Password: userPassword,
		Email:    userEmail,
	}

	// 创建客户端
	c := u.userConn.NewUserClient()

	// 设置超时时间 | set the timeout
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发起Rpc请求
	resp, err := c.Register(rpcCtx, &registerReq)

	if err != nil {
		u.logger.Error("register internal error:", zap.String("detail", err.Error()))
		ctx.JSON(http.StatusInternalServerError, rsp.BaseRsp{
			StatusCode: rsp.Internal,
		})
		return
	}

	// 根据不同的情况返回结果
	switch resp.Case {
	case pb.RegisterResp_OK:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.OK,
			StatusMsg:  "success",
		})
	case pb.RegisterResp_EmailRegistered:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.EmailRegistered,
		})
	case pb.RegisterResp_NameRegistered:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.NameRegistered,
		})
	}
}

// RegisterVerify 验证用户邮箱
func (u *UserHandler) RegisterVerify(ctx *gin.Context) {
	// 创建客户端
	c := u.userConn.NewUserClient()

	// 取得url中的taskId参数，提供给rpc接口进行信息查找
	taskId := ctx.Param("task_id")

	// 设置超时时间 | set the timeout
	rpcCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 发起Rpc请求
	resp, err := c.RegisterVerify(rpcCtx, &pb.RegisterVerifyReq{
		TaskId: taskId,
	})
	if err != nil {
		u.logger.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, rsp.BaseRsp{
			StatusCode: rsp.Internal,
		})
		return
	}
	// 包装并返回结果
	ctx.JSON(http.StatusOK, rsp.TokenRsp{
		BaseRsp: rsp.BaseRsp{
			StatusCode: rsp.OK,
			Data: model.UserVO{
				Name:   resp.User.Name,
				Email:  resp.User.Email,
				Avatar: resp.User.Avatar,
			},
		},
		Token: resp.Token,
	})
}

// UserLogin 用户登录
func (u *UserHandler) UserLogin(ctx *gin.Context) {
	// 接受form表单信息
	// 用户提供的信息有四种情况 因为用户登录可以使用用户名或者邮箱 证明自己就是该用户可以使用密码或邮箱验证码
	userName := ctx.PostForm("name")
	userEmail := ctx.PostForm("email")
	userPassword := ctx.PostForm("password")
	userVerificationCode := ctx.PostForm("verification_code")
	// 用于请求的信息
	loginReq := pb.LoginReq{}
	if userName != "" {
		loginReq.LoginBy = &pb.LoginReq_Name{
			Name: userName,
		}
	} else if userEmail != "" {
		loginReq.LoginBy = &pb.LoginReq_Email{
			Email: userEmail,
		}
	} else {
		ctx.JSON(http.StatusBadRequest, rsp.BaseRsp{
			StatusCode: rsp.EmptyParam,
		})
		return
	}
	if userPassword != "" {
		loginReq.CertifyBy = &pb.LoginReq_Password{
			Password: userPassword,
		}
	} else if userVerificationCode != "" {
		loginReq.CertifyBy = &pb.LoginReq_VerificationCode{
			VerificationCode: userVerificationCode,
		}
	} else {
		ctx.JSON(http.StatusBadRequest, rsp.BaseRsp{
			StatusCode: rsp.EmptyParam,
		})
		return
	}

	//创建客户端
	c := u.userConn.NewUserClient()

	// 设置超时时间
	rpcCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//发起 Rpc请求
	resp, err := c.Login(rpcCtx, &loginReq)

	// 处理服务内部错误
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, rsp.BaseRsp{
			StatusCode: rsp.Internal,
			StatusMsg:  err.Error(),
		})
		return
	}

	switch resp.Case {
	case pb.LoginResp_OK:
		// 包装并返回jwt令牌信息
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.OK,
			Data: model.UserVoWithToken{
				UserVO: model.UserVO{
					Name:   resp.GetUser().Name,
					Email:  resp.GetUser().Email,
					Avatar: resp.GetUser().Avatar,
				},
				Token: resp.Token,
			},
		})
	case pb.LoginResp_PasswordWrong:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.PasswordWrong,
		})
	case pb.LoginResp_UserNotExists:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.UserNotExists,
		})
	case pb.LoginResp_VerificationCodeInvalid:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.VerificationCodeInvalid,
		})
	}
}

// GetVerificationCode 获取登录验证码
func (u *UserHandler) GetVerificationCode(ctx *gin.Context) {
	// 获取前端传来的数据
	email := ctx.Query("email")
	name := ctx.Query("name")
	// 包装请求数据
	req := &pb.GetVerificationCodeReq{}
	if email != "" {
		req.GetBy = &pb.GetVerificationCodeReq_Email{
			Email: email,
		}
	} else if name != "" {
		req.GetBy = &pb.GetVerificationCodeReq_Name{
			Name: name,
		}
	} else {
		ctx.JSON(http.StatusBadRequest, rsp.BaseRsp{
			StatusCode: rsp.EmptyParam,
		})
		return
	}
	// 从gin.Context中取出UserClient实例
	val, ok := ctx.Get(UserClientKey)
	if !ok {
		u.logger.Error("user client not found")
		ctx.JSON(http.StatusInternalServerError, rsp.BaseRsp{
			StatusCode: rsp.Internal,
		})
		return
	}
	c := val.(pb.UserClient)
	// 设置超时时间
	// set the timeout
	rpcCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// 发起请求
	// Send Rpc request to the server
	resp, err := c.GetVerificationCode(rpcCtx, req)
	if err != nil {
		u.logger.Error("failed to get verification code", zap.String("detail", err.Error()))
		ctx.JSON(http.StatusInternalServerError, rsp.BaseRsp{
			StatusCode: rsp.Internal,
		})
		return
	}
	// 根据业务情况处理
	// Handle according to the business situation
	switch resp.GetCase() {
	case pb.GetVerificationCodeResp_OK:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.OK,
		})
	case pb.GetVerificationCodeResp_UserNotExists:
		ctx.JSON(http.StatusOK, rsp.BaseRsp{
			StatusCode: rsp.UserNotExists,
		})
	}
}
