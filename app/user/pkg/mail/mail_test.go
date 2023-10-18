package mail

import (
	"log"
	"myweb/app/user/internal/conf"
	"os"
	"testing"
)

func TestSendMail(t *testing.T) {
	// 读取模板文件
	body, err := os.ReadFile("D:\\Code\\Project\\GoProject\\myweb\\app\\user\\pkg\\util\\mail\\template\\register.html")
	if err != nil {
		log.Printf("io error")
	}
	NewEmailUtil(conf.NewConfig()).SendMail("2825504436@qq.com", "luo-cheng-xi邮箱验证", string(body))
}
