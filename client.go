package wail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type SenderConfig struct {
	// Name that will be displayed above emails
	Name     string
	Login    string
	Password string
}

type SmtpConfig struct {
	Host       string
	Port       uint16
	UseAuth    bool
	Sender     SenderConfig
	TlsConfig  *tls.Config
	// PreferAuth smtp.Auth
	// UseTls bool

	maxMsgSize uint
}

type SmtpClient struct {
	cfg    *SmtpConfig
	client *smtp.Client
}

// Notes:
// Sender login and mail from address may be different?
// Gmail uses 465 port for ssl connection and 587 for tls/starttls

func NewClient(cfg *SmtpConfig) *SmtpClient {
	return &SmtpClient{cfg: cfg}
}

func (s *SmtpClient) Dial() error {
	address := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return err
	}

	// TODO: if tls

	conn = tls.Client(conn, s.cfg.TlsConfig)

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}

	s.client = c

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	if err := c.Hello(hostname); err != nil {
		return err
	}

	if ok, value := c.Extension("SIZE"); ok {
		size, err := strconv.Atoi(value)
		if err != nil {
			s.cfg.maxMsgSize = uint(size)
		} else {
			s.cfg.maxMsgSize = 10485760; // 10 MB
		}
	}

	// if UseTls
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(s.cfg.TlsConfig); err != nil {
			c.Quit()
			return err
		}
	}

	if s.cfg.UseAuth {
		if s.cfg.Sender.Login == "" {
			return errors.New("wail: sender login doesn't specified")
		}

		if s.cfg.Sender.Password == "" {
			return errors.New("wail: sender password doesn't specified")
		}

		// auth := s.cfg.PreferAuth
		var auth smtp.Auth = nil

		// if auth == nil {
			if ok, authMethod := c.Extension("AUTH"); ok {
				switch {
				case strings.Contains(authMethod, "LOGIN"):
					auth = LoginAuth(s.cfg.Sender.Login, s.cfg.Sender.Password)
				case strings.Contains(authMethod, "CRAM-MD5"):
					auth = smtp.CRAMMD5Auth(s.cfg.Sender.Login, s.cfg.Sender.Password)
				case strings.Contains(authMethod, "XOAUTH2"):
					{
						// var token oauth2.Token
						// token.
						// 	a = XoAuth2Auth(s.cfg.Sender.Login)
					}
				case strings.Contains(authMethod, "PLAIN"):
					auth = smtp.PlainAuth("", s.cfg.Sender.Login, s.cfg.Sender.Password, s.cfg.Host)
				}

				if auth == nil {
					c.Quit()
					return errors.New("wail: can't retrieve authentication method")
				}
			// }
		}

		if err := c.Auth(auth); err != nil {
			c.Quit()
			return err
		}
	}

	return nil
}

func (s *SmtpClient) Close() error {
	return s.client.Quit()
}

// func (s *SmtpClient) reconnect() error {
// 	return nil
// }

func (s *SmtpClient) Send(m *Mail) error {
	// TODO: reconnect if connection was closed
	if err := s.client.Noop(); err != nil {
		// reconnect
	}

	if err := s.client.Mail(s.cfg.Sender.Login); err != nil {
		return err
	}

	if len(m.recipients) == 0 {
		return errors.New("wail: no recipients provided to send email")
	}

	for _, email := range m.recipients {
		// for _, email := range emails {
			if err := s.client.Rcpt(email); err != nil {
				return err
			}
		// }
	}

	m.mb.SetFieldFrom(s.cfg.Sender.Name, s.cfg.Sender.Login)
	header := m.mb.GetResultMessage(s.cfg.maxMsgSize)

	log.Println("Header:", string(header))

	w, err := s.client.Data()
	if err != nil {
		return nil
	}

	_, err = w.Write(header)
	if err != nil {
		w.Close()
		return err
	}

	return w.Close()
	// return nil
}
