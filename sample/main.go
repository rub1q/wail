package main

import (
	"log"
	"os"
	"time"

	"github.com/rub1q/wail"
)

func main() {
	cfg := &wail.SmtpConfig{
		Server: wail.ServerConfig{
			Host:           "smtp.mail.ru",
			Port:           465,
			NeedAuth:       true,
			ConnectTimeout: 10 * time.Second,
		},
		Sender: wail.SenderConfig{
			Name:     "Alex",
			Login:    os.Getenv("SENDER_LOGIN"),
			Password: os.Getenv("SENDER_PWD"),
		},
	}

	c := wail.NewClient(cfg)

	err := c.Dial()
	if err != nil {
		log.Fatal(err.Error())
	}

	defer c.Close()

	mailCfg := &wail.MailConfig{
		Charset: wail.UTF8,
		Encoding: wail.Base64,
	}

	// You can pass nil config. In that case will be used a default one
	mail := wail.NewMail(mailCfg)

	mail.SetSubject("Test subject")

	err = mail.To("example@example.com")
	if err != nil {
		log.Fatal(err.Error())
	}

	mt, err := CreateMultipartMixedMessage()
	if err != nil {
		log.Fatal(err.Error())	
	}

	mail.SetMessage(&mt)

	err = c.Send(mail)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Message sent successfully")
}

// CreatePlainTextMessage creates a text/plain message
func CreatePlainTextMessage() wail.TextMessage {
	mt := wail.NewTextMessage()
	mt.Set(wail.TextPlain, []byte("Hello, World"))

	return mt
}

// CreateHtmlTextMessage creates a text/html message
func CreateHtmlTextMessage() wail.TextMessage {
	mt := wail.NewTextMessage()
	mt.Set(wail.TextHtml, []byte("<b>Hello, World</b>"))

	return mt
}

// CreateMultipartMixedMessage creates a multipart/mixed message
func CreateMultipartMixedMessage() (wail.MultipartMixedMessage, error) {
	mt := wail.NewMultipartMixedMessage()

	mt.SetText(wail.TextPlain, []byte("Hello, World"))

	a1 := wail.NewAttachment()

	err := a1.ReadFromFile("D:\\1.csv")
	if err != nil {
		return wail.MultipartMixedMessage{}, err
	}

	a2 := wail.NewAttachment()

	err = a2.ReadFromFile("D:\\pictures\\4fzj5r9pFbY.jpg")
	if err != nil {
		return wail.MultipartMixedMessage{}, err
	}

	mt.AddAttachment(a1)
	mt.AddAttachment(a2)

	return mt, nil
}

// CreateMultipartAltMessage creates a multipart/alt message
func CreateMultipartAltMessage() wail.MultipartAltMessage {
	mt := wail.NewMultipartAltMessage()

	mt.SetPlainText([]byte("Hello, World"), 2)
	mt.SetHtmlText([]byte("<b>Hello, World</b>"), 3)

	return mt
}
