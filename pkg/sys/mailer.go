package sys

import (
	"crypto/tls"
	"log"
	"qilin-api/pkg/conf"
	"gopkg.in/gomail.v2"
)

type Mailer interface {
	Send(to, subject, body string) error
}

type MailerImpl struct {
	replyTo		string
	from 		string
	dialer 		*gomail.Dialer
}

func NewMailer(config conf.Mailer) (mailer Mailer) {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}
	mailer = &MailerImpl{config.ReplyTo, config.From, dialer}
	return
}

func (mailer *MailerImpl) Send(to, subject, body string) (err error) {

	m := gomail.NewMessage()
	m.SetHeader("From", mailer.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if mailer.replyTo != "" {
		m.SetHeader("ReplyTo", mailer.replyTo)
	}
	m.SetBody("text/html", body)

	err = mailer.dialer.DialAndSend(m)
	if err != nil {
		log.Println(err)
	}

	return
}