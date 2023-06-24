package wail

import (
	"os"
	"testing"
	"time"
)

func testClientNoConfig() *SmtpClient {
	return NewClient(nil)
}

func testConfigWithAuth() *SmtpClient {
	cfg := &SmtpConfig{
		Server: ServerConfig{
			Host:           "smtp.mail.ru",
			Port:           465,
			NeedAuth:       true,
			ConnectTimeout: 10 * time.Second,
		},
		Sender: SenderConfig{
			Name:     "Test",
			Login:    os.Getenv("SENDER_LOGIN"),
			Password: os.Getenv("SENDER_PWD"),
		},
	}

	return NewClient(cfg)
}

func testConfigEmptyLoginPassword() *SmtpClient {
	cfg := &SmtpConfig{
		Server: ServerConfig{
			Host:     "smtp.mail.ru",
			Port:     465,
			NeedAuth: true,
		},
		Sender: SenderConfig{
			Name: "Alex",
		},
	}

	return NewClient(cfg)
}

func testConfigNoAuth() *SmtpClient {
	cfg := &SmtpConfig{
		Server: ServerConfig{
			Host:     "smtp.mail.ru",
			Port:     465,
			NeedAuth: false,
		},
	}

	return NewClient(cfg)
}

func testConfigNoEncrypt() *SmtpClient {
	cfg := &SmtpConfig{
		Server: ServerConfig{
			Host:           "smtp.mail.ru",
			Port:           465,
			NeedAuth:       false,
			EncryptType:    EncryptNone,
			ConnectTimeout: 10 * time.Second,
		},
	}

	return NewClient(cfg)
}

func testEmptyConfig() *SmtpClient {
	return NewClient(&SmtpConfig{})
}

func TestDial(t *testing.T) {
	if err := testClientNoConfig().Dial(); err == nil {
		t.Error("smtp config should be provided")
	}

	if err := testConfigWithAuth().Dial(); err != nil {
		t.Errorf("testConfigWithAuth test failed: %v", err)
	}

	if err := testConfigEmptyLoginPassword().Dial(); err == nil {
		t.Error("can't do Auth with an empty sender login and password")
	}

	if err := testConfigNoAuth().Dial(); err != nil {
		t.Errorf("testConfigNoAuth test failed: %v", err)
	}

	// testConfigNoEncrypt should fails because
	// the specified server demands an encrypt connection
	if err := testConfigNoEncrypt().Dial(); err == nil {
		t.Error("server expects an encrypt connection")
	}

	if err := testEmptyConfig().Dial(); err == nil {
		t.Error("config should not be empty")
	}
}

func TestClose(t *testing.T) {
	// Do Close() before Dial()
	if err := testClientNoConfig().Close(); err == nil {
		t.Error("can't do Close() before Dial()")
	}
}

func TestSend(t *testing.T) {
	// Do Send() before Dial()
	if err := testClientNoConfig().Send(nil); err == nil {
		t.Error("can't do Send() before Dial()")
	}

	// Do Send() a nil mail
	c := testConfigWithAuth()
	c.Dial()

	defer c.Close()

	if err := c.Send(nil); err == nil {
		t.Error("can't send a nil mail")
	}

	// Do Send() -> Close() -> Send()
	// This test checks whether reconnection to the server is being performed
	mail := NewMail(nil)

	mail.SetSubject("тема")
	mail.To("example@example.com")

	mt := NewTextMessage()
	mt.Set(TextPlain, []byte("Hello, World"))

	mail.SetMessage(&mt)

	c.Send(mail)
	c.Close()
	
	if err := c.Send(mail); err != nil {
		t.Error(err)
	}
}
