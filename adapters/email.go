package adapters

import (
	"os"

	"gopkg.in/gomail.v2"
)

func Send_Email(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@demomailtrap.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), 587, os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"))

	return d.DialAndSend(m)
}
