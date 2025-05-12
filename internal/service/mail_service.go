package service

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/go-mail/mail/v2"
)

type MailService interface {
	Send(to, subject, body string) error
}

type mailServiceImpl struct {
	dialer *mail.Dialer
	from   string
}

type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewMailService(cfg MailConfig) MailService {
	d := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         cfg.Host,
	}
	// если нужно, можно переопределить локальное имя
	if hn, err := os.Hostname(); err == nil {
		d.LocalName = hn
	}
	return &mailServiceImpl{
		dialer: d,
		from:   cfg.From,
	}
}

func (m *mailServiceImpl) Send(to, subject, body string) error {
	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("mail send failed: %w", err)
	}
	return nil
}
