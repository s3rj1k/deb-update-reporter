package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

type header struct {
	Name  string
	Value string
}

func getHeaders(headers ...header) (out string) {
	for _, header := range headers {
		if len(header.Value) == 0 || len(header.Name) == 0 {
			continue
		}
		out += fmt.Sprintf(
			"%s: %s\r\n",
			header.Name,
			header.Value,
		)
	}

	return out + "\r\n"
}

// HeadersConfig describes headers configuration.
type HeadersConfig struct {
	From    string `json:"From" yaml:"From"`
	To      string `json:"To,omitempty" yaml:"To,omitempty"`
	Subject string `json:"Subject" yaml:"Subject"`
	ReplyTo string `json:"ReplyTo,omitempty" yaml:"ReplyTo,omitempty"`
}

// SMTPConfig describes SMTP configuration.
type SMTPConfig struct {
	Server   string `json:"Server" yaml:"Server"`
	Port     int    `json:"Port" yaml:"Port"`
	Address  string `json:"Address" yaml:"Address"`
	Password string `json:"Password" yaml:"Password"`
}

// MailConfig describes SMTP config.
type MailConfig struct {
	SMTP    SMTPConfig    `json:"SMTP" yaml:"SMTP"`
	Headers HeadersConfig `json:"Headers" yaml:"Headers"`
	Body    string        `json:"-" yaml:"-"`
}

// SendHTML sends HTML email via SMTP.
func (c *MailConfig) SendHTML() error {
	headers := getHeaders(
		header{"From", c.Headers.From},
		header{"To", c.Headers.To},
		header{"Subject", c.Headers.Subject},
		header{"Reply-To", c.Headers.ReplyTo},
		header{"MIME-Version", "1.0"},
		header{"Content-Type", `text/html; charset="utf-8"`},
		header{"Content-Transfer-Encoding", "base64"},
	)

	body := headers + base64.StdEncoding.EncodeToString([]byte(c.Body))

	return smtp.SendMail(
		net.JoinHostPort(
			c.SMTP.Server,
			strconv.Itoa(c.SMTP.Port),
		),
		smtp.PlainAuth(
			"",
			c.SMTP.Address,
			c.SMTP.Password,
			c.SMTP.Server,
		),
		c.SMTP.Address,
		[]string{c.Headers.To},
		[]byte(body),
	)
}

// SendText sends text email via SMTP.
func (c *MailConfig) SendText() error {
	headers := getHeaders(
		header{"From", c.Headers.From},
		header{"To", c.Headers.To},
		header{"Subject", c.Headers.Subject},
		header{"Reply-To", c.Headers.ReplyTo},
		header{"MIME-Version", "1.0"},
		header{"Content-Type", `text/plain; charset="utf-8"`},
		header{"Content-Transfer-Encoding", "base64"},
	)

	body := headers + base64.StdEncoding.EncodeToString([]byte(c.Body))

	return smtp.SendMail(
		net.JoinHostPort(
			c.SMTP.Server,
			strconv.Itoa(c.SMTP.Port),
		),
		smtp.PlainAuth(
			"",
			c.SMTP.Address,
			c.SMTP.Password,
			c.SMTP.Server,
		),
		c.SMTP.Address,
		[]string{c.Headers.To},
		[]byte(body),
	)
}
