package channels

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/bytedance/sonic"
	"github.com/forhsd/alert/errors"
	"github.com/matcornic/hermes/v2"
	"github.com/spf13/cast"
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

func (e *EmailChannel) Send(ctx context.Context, title string, content []*errors.ErrorDetail) error {
	email := e.generateHermesEmail(title, content)
	return e.sendSMTPEmail(ctx, title, email)
}

func (e *EmailChannel) Close() error {
	return nil
}

func (e *EmailChannel) Name() string {
	return e.BaseChannel.Name
}

// generateHermesEmail 将告警内容转换为 Hermes 邮件结构
func (e *EmailChannel) generateHermesEmail(title string, errors []*errors.ErrorDetail) hermes.Email {

	var data [][]hermes.Entry
	for i, err := range errors {
		metadata, _ := sonic.MarshalIndent(err.Metadata, "   ", "  ")
		row := []hermes.Entry{
			{Key: "index", Value: cast.ToString(i + 1)},
			// {Key: "ID", Value: err.ID},
			{Key: "Message", Value: err.Message},
			// {Key: "Stack", Value: err.Stack},
			{Key: "Level", Value: err.Level.String()},
			{Key: "Count", Value: cast.ToString(err.Count)},
			{Key: "FirstSeen", Value: err.FirstSeen.Format(time.TimeOnly)},
			{Key: "LastSeen", Value: err.LastSeen.Format(time.TimeOnly)},
			{Key: "Metadata", Value: string(metadata)},
		}
		data = append(data, row)
	}

	return hermes.Email{
		Body: hermes.Body{
			Name: "系统管理员", // 收件人称呼
			Intros: []string{
				fmt.Sprintf("系统告警通知：%s", title),
				"以下是过去一段时间内系统的错误报告汇总：",
			},
			Dictionary: []hermes.Entry{
				{Key: "报告时间", Value: time.Now().Format("2006-01-02 15:04:05")},
				{Key: "告警级别", Value: "系统错误汇总"},
			},
			Actions: []hermes.Action{
				{
					Instructions: "请及时登录系统查看详情并处理",
					Button: hermes.Button{
						Color: "#22BC66", // 红色按钮，突出告警
						Text:  "登录管理后台",
						Link:  "https://holder-mb.holderzone.com/web",
					},
				},
			},

			Table: hermes.Table{
				Data: data,
				// Columns: hermes.Columns{
				// 	CustomWidth: map[string]string{"index": "15%", "Content": "80%"},
				// 	CustomAlignment: map[string]string{"Content": "left"},
				// },
			},

			// 将原始告警内容放在 Outros 部分
			Outros: []string{
				"此邮件为自动发送，请勿直接回复。",
			},
		},
	}
}

// sendSMTPEmail 通过 SMTP 发送邮件
func (e *EmailChannel) sendSMTPEmail(ctx context.Context, subject string, email hermes.Email) error {

	_ = ctx
	// // 设置超时上下文
	// ctx, cancel := context.WithTimeout(ctx, e.timeout)
	// defer cancel()

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

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)
	// m.Attach("/home/Alex/lolcat.jpg")

	c := e.EmailConfig
	d := gomail.NewDialer(c.SmtpServer, int(c.SmtpPort), c.UserName, c.Password)
	return d.DialAndSend(m)
}
