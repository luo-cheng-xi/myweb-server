package util

import (
	"github.com/golang-jwt/jwt/v5"
	"myweb/app/user/internal/conf"
	"time"
)

type JwtUtil struct {
	signedKey string
}

func NewJwtUtil(c *conf.Config) *JwtUtil {
	return &JwtUtil{
		signedKey: c.JwtConf.JwtSignedKey,
	}
}

type Payload struct {
	ID int
	jwt.RegisteredClaims
}

// GetJwt 通过用户id参数得到JWT令牌
func (rx *JwtUtil) GetJwt(id int) string {
	claims := Payload{
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 过期时间24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                     // 生效时间
		},
	}
	// 使用HS256签名算法
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(rx.signedKey))
	if err != nil {
		return ""
	}
	return s
}

// ParseJwt 解析JWT
func (rx *JwtUtil) ParseJwt(tokenString string) (*Payload, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(rx.signedKey), nil
	})
	if err != nil {
		return nil, err
	}
	//检查信息是否为payload类型，如果是的话就返回其中包含的信息
	if claims, ok := t.Claims.(*Payload); ok && t.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

// GetUserIdFromJwt 从Jwt令牌中获取用户ID信息
func (rx *JwtUtil) GetUserIdFromJwt(tokenString string) (int, error) {
	load, err := rx.ParseJwt(tokenString)
	if err != nil {
		return 0, err
	}
	return load.ID, nil
}
