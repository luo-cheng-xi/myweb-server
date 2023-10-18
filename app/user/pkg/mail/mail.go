package mail

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
	"log"
	"myweb/app/user/internal/conf"
)

type EmailUtil struct {
	From    string
	SMTPKey string
}

func NewEmailUtil(config *conf.Config) *EmailUtil {
	return &EmailUtil{
		From:    config.EmailConf.From,
		SMTPKey: config.EmailConf.SMTPKey,
	}
}

// SendMail 发送邮件，我使用了text/html格式的body内容
func (e EmailUtil) SendMail(to string, header string, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From) // 发件人
	m.SetHeader("To", to)       // 收件人
	//m.SetAddressHeader("Cc", "xi_shu_ba_wang@qq.com", "target") //抄送
	m.SetHeader("Subject", header)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.qq.com", 25, "lcx-test@qq.com", e.SMTPKey)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Fatal(err)
	}
}
