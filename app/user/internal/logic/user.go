package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"html/template"
	"log"
	"math/rand"
	"myweb/app/user/internal/cache"
	"myweb/app/user/internal/conf"
	"myweb/app/user/internal/dao"
	"myweb/app/user/internal/model/dto"
	"myweb/app/user/internal/model/po"
	"myweb/app/user/pkg/mail"
	"myweb/app/user/pkg/util"
	"os"
	"strconv"
	"time"
)

var (
	ErrRegisterVerificationLinkInvalid = errors.New("err register verification link invalid")
	ErrNameRegistered                  = errors.New("name registered")
	ErrEmailRegistered                 = errors.New("email registered")
)

type UserLogic struct {
	userDao          *dao.UserDao
	cache            *cache.Cache
	jwtUtil          *util.JwtUtil
	mailUtil         *mail.EmailUtil
	logger           *zap.Logger
	conf             *conf.Config
	verifyLinkPrefix string
}

func NewUserLogic(
	userDao *dao.UserDao,
	cache *cache.Cache,
	jwtUtil *util.JwtUtil,
	mailUtil *mail.EmailUtil,
	logger *zap.Logger,
	conf *conf.Config) *UserLogic {
	return &UserLogic{
		userDao:          userDao,
		jwtUtil:          jwtUtil,
		cache:            cache,
		mailUtil:         mailUtil,
		logger:           logger,
		verifyLinkPrefix: conf.RegisterConf.VerifyLinkPrefix, //注入验证链接前缀变量，用于生成邮箱验证码
	}
}
func (s *UserLogic) GetVerificationCode(email string) error {
	// 生成验证码
	// Generate verification code
	verificationCode := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	// 读取模板文件
	// Read the template file
	tmpl, err := os.ReadFile("D:\\Code\\Project\\GoProject\\myweb\\app\\user\\pkg\\mail\\template\\verification_code.html")
	if err != nil {
		return err
	}
	tmplStr := string(tmpl)
	// 向模板文件插入验证码
	// Insert verification code into the template file
	buf := new(bytes.Buffer)
	data := struct{ VerificationCode string }{VerificationCode: verificationCode}
	parsed, err := template.New("verificationMail").Parse(tmplStr)
	if err != nil {
		return err
	}
	if err = parsed.Execute(buf, data); err != nil {
		return err
	}
	// 向目标邮箱发送验证码
	// Send the verification code to the target mailbox
	s.mailUtil.SendMail(email, "luo-cheng-xi个人网站登录验证码", buf.String())
	// 在redis中存储验证码，用于后续的校验工作,存储时间为15分钟
	// The verification code is stored in redis for subsequent verification for 15 minutes
	err = s.cache.Set(context.Background(), cache.CacheLoginVerificationCodePrefix+email, verificationCode, 15*time.Minute)
	if err != nil {
		return err
	}
	return nil
}

/*
Register 注册业务逻辑
ErrNameRegistered表示用户名已经被注册
ErrEmailRegistered表示邮箱已经被注册
*/
func (s *UserLogic) Register(user dto.RegisterUserDto) error {
	// 检查用户名是否已经被占用
	// Check whether the name has been registered
	if s.userDao.ContainsName(user.Name) {
		log.Printf("name registered")
		return ErrNameRegistered
	}
	// 检查邮箱是否已经注册过
	// Check whether the email address has been registered
	if s.userDao.ContainsEmail(user.Email) {
		log.Printf("email registered")
		return ErrEmailRegistered
	}
	// 新建一个uuid生成器
	// Create a UUID generator
	generate, err := util.NewIDGenerator(1, 1)
	if err != nil {
		return err
	}
	taskId := strconv.Itoa(int(generate.GetNextID()))
	// 读取模板文件
	// Read the template file
	tmpl, err := os.ReadFile("D:\\Code\\Project\\GoProject\\myweb\\app\\user\\pkg\\mail\\template\\register.html")
	if err != nil {
		return err
	}
	tmplStr := string(tmpl)
	s.logger.Debug(tmplStr)
	// 插入链接到模板中
	// Insert the link to the template
	data := struct{ VerificationLink string }{VerificationLink: s.verifyLinkPrefix + taskId}
	parsed, err := template.New("verificationLinkMail").Parse(tmplStr)
	buf := new(bytes.Buffer)
	if err = parsed.Execute(buf, data); err != nil {
		return err
	}
	// 发送邮箱验证码给用户
	// Send verification code to user
	s.logger.Debug(buf.String())
	s.mailUtil.SendMail(user.Email, "luo-cheng-xi邮箱验证", buf.String())
	// 包装用户信息
	// Packaging user information
	userInfo := po.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}
	// 在redis中存储用户信息，保留30分钟
	// Store user information in redis for 30 minutes
	err = s.cache.Set(context.Background(), cache.CacheRegisterVerificationTaskPrefix+taskId, userInfo, time.Minute*30)
	s.logger.Debug("存储用户信息", zap.String("redis key : ", cache.CacheRegisterVerificationTaskPrefix+taskId))
	if err != nil {
		return err
	}
	return err
}

func (s *UserLogic) RegisterVerify(taskId string) (*dto.UserWithJwt, error) {
	key := cache.CacheRegisterVerificationTaskPrefix + taskId
	userInfo := po.User{}
	ok, err := s.cache.Get(context.Background(), key, &userInfo)
	if err != nil {
		s.logger.Error("failed to get from cache", zap.String("detail", err.Error()))
		return nil, err
	}
	if !ok {
		return nil, ErrRegisterVerificationLinkInvalid
	}
	// 从redis中删除缓存
	err = s.cache.Delete(context.Background(), key)
	if err != nil {
		s.logger.Error("failed to delete from cache", zap.String("detail", err.Error()))
		return nil, err
	}
	// 加密用户密码 | encrypt user's password
	encryptedPassword, err := util.EncryptPassword(userInfo.Password)
	if err != nil {
		s.logger.Error("failed to encrypt user's password", zap.String("detail", err.Error()))
		return nil, err
	}
	// 将加密后的密码存入userInfo实例
	userInfo.Password = encryptedPassword
	// 存储用户信息
	s.userDao.CreateUser(&userInfo)
	// 生成jwt令牌
	token := s.jwtUtil.GetJwt(int(userInfo.ID))
	return &dto.UserWithJwt{
		Name:   userInfo.Name,
		Email:  userInfo.Email,
		Avatar: userInfo.Avatar,
		Token:  token,
	}, nil
}
