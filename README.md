# Wail
* [Overview](#overview)
* [Installation](#installation)
* [Usage](#usage)
  * [Step #1. Creating a SMTP config](#step-1-creating-a-smtp-config)
  * [Step #2. Creating a client and establishing a connection]([#step-2-creating-a-client-and-establishing-a-connection)
  * [Step #3. Creating an email](#step-3-creating-an-email)
  * [Step #4. Creating a message](#step-4-creating-a-message)
  * [Step #5. Sending the email](#step-5-sending-the-email)

## Overview
Wail is a simple package to send emails. It supports: 
* SSL/TLS encryption
* Base64 and quoted printable encoding
* Authentication on server
* Plain text and HTML messages
* Attachments

## Installation 
To install the package run the following command:

```
go get -u github.com/rub1q/wail
```
## Usage 
### Step #1. Creating a SMTP config

First things first you need to define a SMTP config. For example as follows:
```Go
cfg := &wail.SmtpConfig{
  Server: wail.ServerConfig{
    Host:           "smtp.example.com",
    Port:           465,
    NeedAuth:       true,
    ConnectTimeout: 10 * time.Second,
    EncryptType:    wail.EncryptSSL,
  },
  Sender: wail.SenderConfig{
    Name:     "Alex",
    Login:    os.Getenv("SENDER_LOGIN"),
    Password: os.Getenv("SENDER_PWD"),
  },
  TlsConfig: &tls.Config{
    InsecureSkipVerify: true,
  },
}
```
A few words about config

By default is using `EncryptType = EncryptSSL`. If the SMTP server supports a `STARTTLS` extension you can change the `EncryptType` value to `EncryptTLS`

The sender's `Name` is using to show it above your emails. If you do not need an authentication set the `NeedAuth` to `false`. Hence, the sender's `Login` and `Password` could be omitted

`TlsConfig` is using to specify additional setting for establishing encrypted connection with the SMTP server

> Note: leave the default `TlsConfig` value if you do not know how to use it

### Step #2. Creating a client and establishing a connection

After the config is done you need to create a new client. Call `NewClient()` and pass it your config: 

```Go
c := wail.NewClient(cfg)
```

Then call `Dial()` to establish a connection with the server. `Dial()` could return an error if something go wrong
```Go
err := c.Dial()
if err != nil {
  log.Fatal(err.Error())
}

defer c.Close()
```

### Step #3. Creating an email
The next step is creating an email object. Call the `NewMail()` method to do it

```Go
mailCfg := &wail.MailConfig{
  Charset:  wail.UTF8,
  Encoding: wail.Base64,
}

mail := wail.NewMail(mailCfg)
```
`NewMail()` is accepting `MailConfig` structure as a parameter. You can pass `nil` if you want to use a default config values

Call `SetSubject()` to set the email subject:
```Go
mail.SetSubject("Test subject")
```

Then specify all recipients by using `To()`, `CopyTo()` and `BlindCopyTo()` methods. For example:
```Go
err = mail.To("example1@example.com", "example2@example.com")
if err != nil {
  log.Fatal(err.Error())
}

err = mail.CopyTo("example3@example.com")
if err != nil {
  log.Fatal(err.Error())
}
```

All three methods could return an error

### Step #4. Creating a message

At the momemt Wail supports 4 message [Content-Types](https://en.wikipedia.org/wiki/MIME): 
* `text/plain`
* `text/html`
* `multipart/mixed`
* `multipart/alternative`

For each Content-Type methods and structures are provided 

Assume you want to send a `text\plain` message. Call the `NewTextMessage()` method: 
```Go
mt := wail.NewTextMessage()
```

And then provide a text you want to send: 
```Go
mt.Set(wail.TextPlain, []byte("Hello, World"))
```

`Set()` accepts `TextPlain` or `TextHtml` as the first parameter. Use the last one if you need to send a HTML message

### Step #5. Sending the email

After you have created the message call `SetMessage()` and pass it your message object.

```Go
mail.SetMessage(&mt)
```

Finally, call `Send()` with an email object argument to send an email:

```Go
err = c.Send(mail)
if err != nil {
  log.Fatal(err.Error())
}
```

