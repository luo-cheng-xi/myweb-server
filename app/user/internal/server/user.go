package server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"myweb/app/api/pb"
	"myweb/app/user/internal/cache"
	"myweb/app/user/internal/conf"
	"myweb/app/user/internal/dao"
	"myweb/app/user/internal/logic"
	"myweb/app/user/internal/model/dto"
	"myweb/app/user/internal/model/po"
	"myweb/app/user/pkg/mail"
	"myweb/app/user/pkg/util"
)

type UserServer struct {
	userLogic                  *logic.UserLogic
	userDao                    *dao.UserDao
	cache                      *cache.Cache
	jwtUtil                    *util.JwtUtil
	mailUtil                   *mail.EmailUtil
	logger                     *zap.Logger
	verifyLinkPrefix           string
	pb.UnimplementedUserServer // 必须要实现这个接口 | must implement this interface
}

// NewUserServer 新建用户服务 | initial user service
func NewUserServer(
	userLogic *logic.UserLogic,
	userDao *dao.UserDao,
	cache *cache.Cache,
	jwtUtil *util.JwtUtil,
	mailUtil *mail.EmailUtil,
	logger *zap.Logger,
	conf *conf.Config) *UserServer {
	return &UserServer{
		userLogic:        userLogic,
		userDao:          userDao,
		jwtUtil:          jwtUtil,
		cache:            cache,
		mailUtil:         mailUtil,
		logger:           logger,
		verifyLinkPrefix: conf.RegisterConf.VerifyLinkPrefix, //注入验证链接前缀变量，用于生成邮箱验证码
	}
}

func (s *UserServer) Login(ctx context.Context, in *pb.LoginReq) (*pb.LoginResp, error) {
	s.logger.Debug("login function was called")
	userInfo := po.User{}
	switch in.LoginBy.(type) {
	case *pb.LoginReq_Name:
		userInfo.Name = in.GetName()
	case *pb.LoginReq_Email:
		userInfo.Email = in.GetEmail()
	default:
		return nil, status.Error(codes.FailedPrecondition, "login param invalid")
	}

	// 先查有没有这个用户
	ok := s.userDao.GetUser(&userInfo)
	if !ok {
		return &pb.LoginResp{Case: pb.LoginResp_UserNotExists}, nil
	}
	// 然后根据用户提供的密码或者邮箱验证码进行校验工作
	switch in.CertifyBy.(type) {
	case *pb.LoginReq_Password:
		// 将所给密码与数据库中的密码比对
		// Compare the given password with the password in the database
		same, err := util.MatchPasswordAndHash(in.GetPassword(), userInfo.Password)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// 处理所给密码和用户密码不符的情况
		// Handle cases where the given password does not match the user password
		if !same {
			return &pb.LoginResp{Case: pb.LoginResp_PasswordWrong}, nil
		}
	case *pb.LoginReq_VerificationCode:
		s.logger.Debug("login by verification code", zap.String("key", cache.CacheLoginVerificationCodePrefix+in.GetEmail()))
		key := cache.CacheLoginVerificationCodePrefix + userInfo.Email
		verificationCode := ""
		ok, err := s.cache.Get(context.Background(), key, &verificationCode)
		// 处理内部错误
		// Process internal error
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		// 没有找到
		// value not found
		if !ok {
			// 告知调用者没有与所给用户信息相关的验证码
			// Informs the caller that there is no verification code associated with the information given to the user
			return &pb.LoginResp{Case: pb.LoginResp_GivenUserHasNoVerificationCode}, nil
		}
		s.logger.Debug("verification code in redis", zap.String("value:", verificationCode))
		// 检查缓存中的验证码是否和用户输入一致
		// Check whether the verification code in the cache is consistent with the given verification code
		if verificationCode != in.GetVerificationCode() {
			return &pb.LoginResp{Case: pb.LoginResp_VerificationCodeInvalid}, nil
		}
		// 此时已经确定无误，删除redis中缓存的验证码
		err = s.cache.Delete(context.Background(), key)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	default:
		return nil, status.Error(codes.FailedPrecondition, "certify param invalid")
	}
	// 比对确认无误，构造jwt令牌
	// The password is confirmed so construct jwt token
	token := s.jwtUtil.GetJwt(int(userInfo.ID))

	// 返回得到的信息
	// Return the resulting information
	return &pb.LoginResp{
		User: &pb.UserInfo{
			Name:   userInfo.Name,
			Email:  userInfo.Email,
			Avatar: userInfo.Avatar,
		},
		Case:  pb.LoginResp_OK,
		Token: token,
	}, nil
}

// Register 请求服务器发送注册邮箱
func (s *UserServer) Register(ctx context.Context, in *pb.RegisterReq) (*pb.RegisterResp, error) {
	user := dto.RegisterUserDto{
		Email:    in.GetEmail(),
		Name:     in.GetName(),
		Password: in.GetPassword(),
	}

	err := s.userLogic.Register(user)

	if err != nil {
		switch err {
		case logic.ErrNameRegistered:
			return &pb.RegisterResp{
				Case: pb.RegisterResp_NameRegistered,
			}, nil
		case logic.ErrEmailRegistered:
			return &pb.RegisterResp{
				Case: pb.RegisterResp_EmailRegistered,
			}, nil
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	// 完成业务逻辑，返回结果
	return &pb.RegisterResp{
		Case: pb.RegisterResp_OK,
	}, nil
}

// RegisterVerify 验证邮箱
func (s *UserServer) RegisterVerify(ctx context.Context, in *pb.RegisterVerifyReq) (*pb.RegisterVerifyResp, error) {
	// 读取前端传来的验证请求的业务id，会通过这个id生成key从缓存中读取注册时用户提供的信息
	taskId := in.GetTaskId()

	ret, err := s.userLogic.RegisterVerify(taskId)
	if err != nil {
		if errors.Is(err, logic.ErrRegisterVerificationLinkInvalid) {

		}
	}
	return &pb.RegisterVerifyResp{
		User: &pb.UserInfo{
			Name:   ret.Name,
			Email:  ret.Email,
			Avatar: ret.Avatar,
		},
		Token: ret.Token,
	}, nil
}

// GetVerificationCode 获取登录验证码
func (s *UserServer) GetVerificationCode(ctx context.Context, in *pb.GetVerificationCodeReq) (*pb.GetVerificationCodeResp, error) {
	email := ""
	switch in.GetBy.(type) {
	case *pb.GetVerificationCodeReq_Email:
		email = in.GetEmail()
	case *pb.GetVerificationCodeReq_Name:
		userInfo := po.User{
			Name: in.GetName(),
		}
		exist := s.userDao.GetUser(&userInfo)
		if !exist {
			return &pb.GetVerificationCodeResp{
				Case: pb.GetVerificationCodeResp_UserNotExists,
			}, nil
		}
		email = userInfo.Email
	}

	err := s.userLogic.GetVerificationCode(email)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetVerificationCodeResp{
		Case: pb.GetVerificationCodeResp_OK,
	}, nil
}
