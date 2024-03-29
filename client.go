package wail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// SenderConfig contains information about the sender
type SenderConfig struct {
	// Name specified in this field will be displayed above emails
	Name string

	// Login is the email address from which an emails will be sent.
	// It is also will be used for authentication on server (if required)
	Login string

	// Password from your email account. It is used for authentication on server
	Password string
}

type encryption int

const (
	// EncryptSSL is a default encryption type. Use this
	// type if you want to encrypt your connection but
	// you don't need to call STARTTLS command.
	//
	// This encryption type may be used if you establishing
	// a connection on port 465
	EncryptSSL encryption = iota

	// EncryptTLS encryption type is used if you want to
	// encrypt connection by calling STARTTLS command.
	//
	// This encryption type may be used if you establishing
	// a connection on port 587 or 25 (the last one is not
	// recommended to use)
	EncryptTLS

	// No encryption
	EncryptNone
)

// ServerConfig contains information about the SMTP server
type ServerConfig struct {
	// Host represents the SMTP server address
	Host string

	// Port represents the SMTP server port
	Port uint16

	ConnectTimeout time.Duration

	// NeedAuth is used to indicate that the server
	// demands an authentication before sending emails
	NeedAuth bool

	// EncryptType is an encryption type (SSL, TLS or none)
	EncryptType encryption

	// maxMsgSize is a maximum message size that can be sent to the server.
	// This field is set only if the server returns the SIZE extension
	maxMsgSize uint
}

// SmtpConfig contains information required for establishing connection
// and generating message
type SmtpConfig struct {
	// Server represents the server configuration
	Server ServerConfig

	// Sender represents the sender configuration
	Sender SenderConfig

	// TlsConfig is the TLS configuration used for TLS or SSL connections.
	//
	// Note: leave the default value if you don't know how to use it
	TlsConfig *tls.Config
}

// SmtpClient represents a client that negotiate with the server
type SmtpClient struct {
	cfg    *SmtpConfig
	client *smtp.Client
}

// NewClient returns the new SMTP client
func NewClient(cfg *SmtpConfig) *SmtpClient {
	return &SmtpClient{cfg: cfg, client: nil}
}

// Dial establishes a connection with the server using
// parameters from SMTP config. If an error occurs
// during a connection Dial will return it
func (s *SmtpClient) Dial() error {
	if s.cfg == nil {
		return errors.New("wail: smtp config is not provided")
	}

	address := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	conn, err := net.DialTimeout("tcp", address, s.cfg.Server.ConnectTimeout)
	if err != nil {
		return err
	}

	if s.cfg.Server.EncryptType == EncryptSSL || s.cfg.Server.EncryptType == EncryptTLS {
		if s.cfg.TlsConfig == nil {
			s.cfg.TlsConfig = &tls.Config{}
		}

		if !s.cfg.TlsConfig.InsecureSkipVerify {
			s.cfg.TlsConfig.ServerName = s.cfg.Server.Host
		}

		conn = tls.Client(conn, s.cfg.TlsConfig)
	}

	var c *smtp.Client

	if s.cfg.Server.ConnectTimeout != 0 {
		connChan := make(chan error)

		go func() {
			defer close(connChan)

			c, err = smtp.NewClient(conn, s.cfg.Server.Host)
			connChan <- err
		}()

		select {
		case <-time.After(s.cfg.Server.ConnectTimeout):
			return errors.New("wail: connection timeout")
		case err := <-connChan:
			if err != nil {
				return err
			}
		}
	} else {
		c, err = smtp.NewClient(conn, s.cfg.Server.Host)
		if err != nil {
			return err
		}
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
		if size, err := strconv.Atoi(value); err == nil {
			s.cfg.Server.maxMsgSize = uint(size)
		}
	}

	if s.cfg.Server.EncryptType == EncryptTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(s.cfg.TlsConfig); err != nil {
				c.Quit()
				return err
			}
		}
	}

	if s.cfg.Server.NeedAuth {
		if s.cfg.Sender.Login == "" {
			return errors.New("wail: sender login is not specified")
		}

		if s.cfg.Sender.Password == "" {
			return errors.New("wail: sender password is not specified")
		}

		var auth smtp.Auth = nil

		if ok, authMethod := c.Extension("AUTH"); ok {
			switch {
			case strings.Contains(authMethod, "LOGIN"):
				auth = LoginAuth(s.cfg.Sender.Login, s.cfg.Sender.Password)
			case strings.Contains(authMethod, "CRAM-MD5"):
				auth = smtp.CRAMMD5Auth(s.cfg.Sender.Login, s.cfg.Sender.Password)
			case strings.Contains(authMethod, "XOAUTH2"):
				{
					// TODO: make support XOAUTH2 auth?
				}
			case strings.Contains(authMethod, "PLAIN"):
				auth = smtp.PlainAuth("", s.cfg.Sender.Login, s.cfg.Sender.Password, s.cfg.Server.Host)
			}

			if auth == nil {
				c.Quit()
				return errors.New("wail: can't retrieve authentication method")
			}
		}

		if err := c.Auth(auth); err != nil {
			c.Quit()
			return err
		}
	}

	return nil
}

// Close closes a connection with the server by sending the QUIT command
func (s *SmtpClient) Close() error {
	if s.client == nil {
		return errors.New("wail: connection with the smtp server is not established")
	}

	return s.client.Quit()
}

// Send assembles the message and sends it to the server
func (s *SmtpClient) Send(m *Mail) error {
	if s.client == nil {
		return errors.New("wail: connection with the smtp server is not established")
	}

	if m == nil {
		return errors.New("wail: an empty mail object has been provided")
	}

	if err := s.client.Noop(); err != nil {
		if err := s.Dial(); err != nil {
			return fmt.Errorf("wail: an error occured while reconnecting to the server (%s)", err.Error())
		}
	}

	if err := s.client.Mail(s.cfg.Sender.Login); err != nil {
		return err
	}

	if len(m.recipients) == 0 {
		return errors.New("wail: no recipients provided to send email")
	}

	for _, email := range m.recipients {
		if err := s.client.Rcpt(email); err != nil {
			return err
		}
	}

	m.mb.SetFieldFrom(s.cfg.Sender.Name, s.cfg.Sender.Login)

	header, err := m.mb.GetResultMessage(s.cfg.Server.maxMsgSize)
	if err != nil {
		return err
	}

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
}
