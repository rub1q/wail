package wail

import (
	"errors"
	"net/mail"
)

type Encoding string

const (
	QuotedPrintable Encoding = "quoted-printable"
	Base64          Encoding = "base64"
)

type Charset string

const (
	UTF8       Charset = "UTF-8"
	ISO_8859_1 Charset = "ISO-8859-1"
	US_ASCII   Charset = "US-ASCII"
)

type AuthMethod string

const (
	Plain   AuthMethod = "PLAIN"
	Login   AuthMethod = "LOGIN"
	XoAuth2 AuthMethod = "XOAUTH2"
	CramMd5 AuthMethod = "CRAM-MD5"
)

type recipientType int

// const (
// 	to  recipientType = 1 << 1
// 	cc  recipientType = 1 << 2
// 	bcc recipientType = 1 << 3
// )

type Recipients []string //map[recipientType][]string

type MailConfig struct {
	Charset  Charset
	Encoding Encoding
	Auth     AuthMethod
}

type Mail struct {
	cfg *MailConfig
	mb  *mimeBuilder

	recipients Recipients
}

var DefaultMailConfig MailConfig = MailConfig{
	Charset:  UTF8,
	Encoding: Base64,
	Auth:     Login,
}

func NewMail(cfg *MailConfig) *Mail {
	var m *Mail

	if cfg != nil {
		m = &Mail{
			cfg: &MailConfig{
				Charset:  cfg.Charset,
				Encoding: cfg.Encoding,
				Auth:     cfg.Auth,
			},
		}
	} else {
		m = &Mail{cfg: &DefaultMailConfig}
	}

	// TODO: parse config properly

	// log.Println("cfg", cfg)
	// log.Println("Auth", m.cfg.Auth)

	m.mb = newMimeBuilder(m.cfg.Charset, m.cfg.Encoding)
	m.recipients = make(Recipients, 0)

	return m
}

// SetSubject sets an email subject. Subject could be empty
func (m *Mail) SetSubject(subj string) {
	m.mb.SetFieldSubject(subj)
}

func (m *Mail) validateAndAppendEmails(emails []string) error { // rType recipientType, 
	if len(emails) == 0 {
		return errors.New("wail: an empty email address list has been provided")
	}

	for _, email := range emails {
		if len(email) > 254 {
			return errors.New("wail: length of the email address must be less than 254 chars")
		} else if _, err := mail.ParseAddress(email); err != nil {
			return err
		}
	}

	// m.recipients[rType] = emails
	m.recipients = append(m.recipients, emails...)
	return nil
}

// func (m *Mail) From(name, address string) error {
// 	if _, err := mail.ParseAddress(address); err != nil {
// 		return err
// 	}

// 	m.mb.SetFieldFrom(name, address)
// 	return nil
// }

// To sets main email addresses to which an email will be sent
func (m *Mail) To(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil { //to, 
		return err
	}

	m.mb.SetFieldTo(emails...)
	return nil
}

// CopyTo sets email addresses to which an email copy will be sent
func (m *Mail) CopyTo(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil { // cc, 
		return err
	}

	m.mb.SetFieldCc(emails...)
	return nil
}

// BlindCopyTo sets email addresses to which an email blind copy will be sent
func (m *Mail) BlindCopyTo(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil { // bcc, 
		return err
	}

	m.mb.SetFieldBcc(emails...)
	return nil
}

// SetMessage sets an email message
func (m *Mail) SetMessage(msg Message) {
	m.mb.SetMessage(msg)
}
