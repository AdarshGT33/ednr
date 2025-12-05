package handler

import (
	"gopkg.in/gomail.v2"
)

func SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "hello@demomailtrap.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", "3bc87a9fefe896f02d31b474f5488f2c")

	return d.DialAndSend(m)
}
