package conf

import (
	"gopkg.in/ini.v1"
	"log"
	"strconv"
)

type Config struct {
	MysqlConf struct {
		DSN string
	}
	JwtConf struct {
		JwtSignedKey string
	}
	EmailConf struct {
		SMTPKey string
		From    string
	}

	RedisConf struct {
		Addr     string
		Password string
		DB       int
	}

	RegisterConf struct {
		VerifyLinkPrefix string
	}
}

// NewConfig 读取配置文件
func NewConfig() *Config {
	f, err := ini.Load("D:\\Code\\Project\\GoProject\\myweb\\app\\user\\configs\\conf.ini")
	if err != nil {
		log.Fatalf("Load ini file error %#v", err)
	}

	// 解析ini文件，获取mysql数据库连接配置
	userName := f.Section("mysql").Key("username").String()
	password := f.Section("mysql").Key("password").String()
	host := f.Section("mysql").Key("host").String()
	schema := f.Section("mysql").Key("schema").String()
	dsn := userName + ":" + password + "@tcp(" + host + ")/" + schema + "?charset=utf8mb4&parseTime=True&loc=Local"

	// 获取jwt令牌配置
	jwtSignedKey := f.Section("jwt").Key("signedKey").String()

	// 获取邮件配置
	from := f.Section("email").Key("from").String()
	SMTPKey := f.Section("email").Key("SMTPKey").String()

	// redis配置
	addr := f.Section("redis").Key("addr").String()
	redisPassword := f.Section("redis").Key("password").String()
	db, err := strconv.ParseInt(f.Section("redis").Key("DB").String(), 10, 32)

	verifyLinkPrefix := f.Section("register").Key("verifyLinkPrefix").String()
	if err != nil {
		log.Fatalf("ini load error %#v", err)
	}

	return &Config{
		MysqlConf: struct{ DSN string }{DSN: dsn},
		JwtConf:   struct{ JwtSignedKey string }{JwtSignedKey: jwtSignedKey},
		EmailConf: struct {
			SMTPKey string
			From    string
		}{SMTPKey: SMTPKey, From: from},
		RedisConf: struct {
			Addr     string
			Password string
			DB       int
		}{Addr: addr, Password: redisPassword, DB: int(db)},
		RegisterConf: struct{ VerifyLinkPrefix string }{VerifyLinkPrefix: verifyLinkPrefix},
	}
}
