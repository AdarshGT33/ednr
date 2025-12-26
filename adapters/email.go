package adapters

import (
	"os"

	"gopkg.in/gomail.v2"
)

type EmailAdapter struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

func NewEmailAdapter() *EmailAdapter {
	return &EmailAdapter{
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: 587,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (e *EmailAdapter) Send(to, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@demomailtrap.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "status of your placed order")
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(e.SMTPHost, e.SMTPPort, e.Username, e.Password)

	return d.DialAndSend(m)
}
