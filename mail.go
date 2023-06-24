package wail

import (
	"errors"
	"net/mail"
)

type encoding string

const (
	QuotedPrintable encoding = "quoted-printable"
	Base64          encoding = "base64"
)

type charset string

const (
	UTF8       charset = "UTF-8"
	ISO_8859_1 charset = "ISO-8859-1"
	US_ASCII   charset = "US-ASCII"
)

type recipients []string

type MailConfig struct {
	Charset  charset
	Encoding encoding
}

type Mail struct {
	cfg *MailConfig
	mb  *mimeBuilder

	recipients recipients
}
 
var DefaultMailConfig MailConfig = MailConfig{
	Charset:  UTF8,
	Encoding: Base64,
}

func NewMail(cfg *MailConfig) *Mail {
	var m *Mail

	if cfg != nil {
		if cfg.Charset == "" {
			cfg.Charset = UTF8		
		}
		
		if cfg.Encoding == "" {
			cfg.Encoding = QuotedPrintable	
		}
		
		m = &Mail{
			cfg: &MailConfig{
				Charset:  cfg.Charset,
				Encoding: cfg.Encoding,
			},
		}
	} else {
		m = &Mail{cfg: &DefaultMailConfig}
	}

	m.mb = newMimeBuilder(m.cfg.Charset, m.cfg.Encoding)
	m.recipients = make(recipients, 0, 10)

	return m
}

// SetSubject sets an email subject. Subject could be empty
func (m *Mail) SetSubject(subj string) {
	m.mb.SetFieldSubject(subj)
}

func (m *Mail) validateAndAppendEmails(emails []string) error {
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

	m.recipients = append(m.recipients, emails...)
	return nil
}

// To sets main email addresses to which an email will be sent
func (m *Mail) To(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil {
		return err
	}

	m.mb.SetFieldTo(emails...)
	return nil
}

// CopyTo sets email addresses to which an email copy will be sent
func (m *Mail) CopyTo(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil {
		return err
	}

	m.mb.SetFieldCc(emails...)
	return nil
}

// BlindCopyTo sets email addresses to which an email blind copy will be sent
func (m *Mail) BlindCopyTo(emails ...string) error {
	if err := m.validateAndAppendEmails(emails); err != nil {
		return err
	}

	m.mb.SetFieldBcc(emails...)
	return nil
}

// SetMessage sets an email message
func (m *Mail) SetMessage(msg Message) {
	m.mb.SetMessage(msg)
}
