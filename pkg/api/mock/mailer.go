package mock

import "qilin-api/pkg/sys"

type mailer struct {
}

func NewMailer() sys.Mailer {
	return &mailer{}
}

func (mailer) Send(to, subject, body string) error {
	return nil
}
