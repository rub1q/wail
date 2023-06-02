package wail

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"

	"golang.org/x/oauth2"
)

type authLogin struct {
	username string
	password string
}

type authXoAuth2 struct {
	username string
	token    oauth2.TokenSource
}

func LoginAuth(username, password string) smtp.Auth {
	return &authLogin{
		username: username,
		password: password,
	}
}

func (l *authLogin) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("wail: unencrypted connection")
	}

	return "LOGIN", nil, nil
}

func (l *authLogin) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch {
		case bytes.Contains(fromServer, []byte("Username")):
			return []byte(l.username), nil
		case bytes.Contains(fromServer, []byte("Password")):
			return []byte(l.password), nil
		default:
			return nil, errors.New("wail: unknown command from server")
		}
	}

	return nil, nil
}

func XoAuth2Auth(username string, token oauth2.TokenSource) smtp.Auth {
	return &authXoAuth2{
		username: username,
		token:    token,
	}
}

func (x *authXoAuth2) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("wail: unencrypted connection")
	}

	t, err := x.token.Token()
	if err != nil {
		return "", nil, errors.New("wail: failed to get token")
	}

	oauth2 := fmt.Sprintf("user=%v\001auth=%v %v\001\001", x.username, t.Type(), t.AccessToken)
	
	return "XOAUTH2", []byte(oauth2), nil
}

func (x *authXoAuth2) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("wail: unexpected challenge")
	}

	return nil, nil
}
