package gmail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/smtp"

	"microservice-alert-service/alert/infrastructure/configuration"
)

type Sender struct {
	cfg    configuration.Config
	logger *log.Logger
}

func NewSender(cfg configuration.Config, logger *log.Logger) *Sender {
	return &Sender{cfg: cfg, logger: logger}
}

func (s *Sender) Send(ctx context.Context, to string, subject string, body string) error {
	if s.cfg.MailHost == "" || s.cfg.MailUsername == "" || s.cfg.MailPassword == "" {
		return errors.New("mail configuration missing")
	}

	from := s.cfg.MailFrom
	if from == "" {
		from = s.cfg.MailUsername
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.MailHost, s.cfg.MailPort)
	auth := smtp.PlainAuth("", s.cfg.MailUsername, s.cfg.MailPassword, s.cfg.MailHost)

	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s\r\n",
		from,
		to,
		subject,
		body,
	))

	conn, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	tlsConfig := &tls.Config{ServerName: s.cfg.MailHost}
	if err := conn.StartTLS(tlsConfig); err != nil {
		return err
	}

	if err := conn.Auth(auth); err != nil {
		return err
	}

	if err := conn.Mail(from); err != nil {
		return err
	}

	if err := conn.Rcpt(to); err != nil {
		return err
	}

	writer, err := conn.Data()
	if err != nil {
		return err
	}

	if _, err := writer.Write(message); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return conn.Quit()
}
