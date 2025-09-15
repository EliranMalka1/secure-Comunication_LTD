package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

type Mailer struct {
	host string
	port string
	from string
}

func NewMailerFromEnv() (*Mailer, error) {
	h := os.Getenv("SMTP_HOST")
	p := os.Getenv("SMTP_PORT")
	f := os.Getenv("SMTP_FROM")
	if h == "" || p == "" || f == "" {
		return nil, fmt.Errorf("SMTP env missing")
	}
	return &Mailer{host: h, port: p, from: f}, nil
}

func (m *Mailer) Send(to, subject, html string) error {
	addr := net.JoinHostPort(m.host, m.port)

	msg := strings.Builder{}
	msg.WriteString("From: " + m.from + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(html)

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.Mail(m.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg.String())); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

var _ = tls.Config{}
