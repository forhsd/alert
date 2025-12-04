package channels

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/forhsd/alert/errors"
	"gopkg.in/gomail.v2"
)

// EmailChannel 邮件渠道
type EmailChannel struct {
	BaseChannel
	EmailConfig
}

type EmailConfig struct {
	SmtpServer string   `json:"smtpServer"`
	SmtpPort   uint     `json:"smtpPort"`
	UserName   string   `json:"userName"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	To         []string `json:"to"`
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(config EmailChannel) (*EmailChannel, error) {
	return &EmailChannel{
		BaseChannel: BaseChannel{},
	}, nil
}

func (e *EmailChannel) Validate() error {

	if err := e.BaseChannel.Validate(); err != nil {
		return err
	}
	if e.SmtpServer == "" {
		return fmt.Errorf("SmtpServer不能为空")
	}
	if e.SmtpPort == 0 {
		return fmt.Errorf("SmtpPort不能为空")
	}
	if e.UserName == "" {
		return fmt.Errorf("邮件用户名不能为空")
	}
	if e.Password == "" {
		return fmt.Errorf("邮件密码不能为空")
	}
	if e.From == "" {
		return fmt.Errorf("发件人不能为空")
	}
	if len(e.To) == 0 {
		return fmt.Errorf("收件人不能为空")
	}

	return nil
}

func (e *EmailChannel) Send(ctx context.Context, title string, content []*errors.ErrorDetail) error {
	return e.sendSMTPEmail(ctx, title, content)
}

func (e *EmailChannel) Close() error {
	return nil
}

func (e *EmailChannel) Name() string {
	return e.BaseChannel.Name
}

// sendSMTPEmail 通过 SMTP 发送邮件
func (e *EmailChannel) sendSMTPEmail(ctx context.Context, subject string, content []*errors.ErrorDetail) error {

	_ = ctx
	// // 设置超时上下文
	// ctx, cancel := context.WithTimeout(ctx, e.timeout)
	// defer cancel()

	email := e.Email(subject, content)

	htmlBody, err := e.GenerateHTML(email)
	if err != nil {
		return err
	}
	// err = os.WriteFile("preview.html", []byte(htmlBody), 0644)
	// if err != nil {
	// 	return err
	// }

	textBody, err := e.GeneratePlainText(email)
	if err != nil {
		return err
	}

	from := mail.Address{
		Name:    e.EmailConfig.From,
		Address: e.EmailConfig.From,
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from.String())     // holderzone@126.com
	m.SetHeader("To", e.EmailConfig.To...) // "forhsd@qq.com",
	m.SetHeader("Subject", subject)        // "主题"

	// m.SetAddressHeader("Cc", "forhsd@qq.com", "峰")
	// m.Attach("/home/Alex/lolcat.jpg")

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	c := e.EmailConfig
	d := gomail.NewDialer(c.SmtpServer, int(c.SmtpPort), c.UserName, c.Password)
	return d.DialAndSend(m)
}
