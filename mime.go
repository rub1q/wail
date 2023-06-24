package wail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"strings"
	"time"
)

// RFC 5322 2.2.3
const lineLengthLimit = 76

type mimeBuilder struct {
	charset     charset
	encoding    encoding
	encoder     mime.WordEncoder
	contentType contentType
	header      map[string]string
}

func newMimeBuilder(charset charset, encoding encoding) *mimeBuilder {
	mb := &mimeBuilder{
		charset:  charset,
		encoding: encoding,
		header:   make(map[string]string),
	}

	switch encoding {
	case QuotedPrintable:
		mb.encoder = mime.QEncoding
	case Base64:
		mb.encoder = mime.BEncoding
	}

	return mb
}

func (m *mimeBuilder) EncodeHeader(value string) string {
	if len(value) == 0 {
		return value
	}

	out := m.encoder.Encode(string(m.charset), value)

	if len(out) > lineLengthLimit {
		out = splitHeader(out)
	}

	return out
}

func (m *mimeBuilder) EncodeBody(body []byte) string {
	var out string

	switch m.encoding {
	case Base64:
		{
			out = base64Encode(body)
		}
	case QuotedPrintable:
		{
			if m, err := qpEncode(body); err != nil {
				out = string(body)
			} else {
				out = m
			}
		}
	}

	return out
}

func (m *mimeBuilder) SetFieldSubject(subj string) {
	m.header["subject"] = m.EncodeHeader(subj)
}

func (m *mimeBuilder) SetFieldFrom(name string, addr string) {
	if len(name) == 0 {
		m.header["from"] = addr
	} else {
		m.header["from"] = fmt.Sprintf("%s <%s>", m.EncodeHeader(name), addr)
	}
}

func (m *mimeBuilder) SetFieldTo(addr ...string) {
	if len(addr) == 0 {
		return
	}

	m.header["to"] = makeAddrString(addr)
}

func (m *mimeBuilder) SetFieldCc(addr ...string) {
	if len(addr) == 0 {
		return
	}

	m.header["cc"] = makeAddrString(addr)
}

func (m *mimeBuilder) SetFieldBcc(addr ...string) {
	if len(addr) == 0 {
		return
	}

	m.header["bcc"] = makeAddrString(addr)
}

func (m *mimeBuilder) SetMessage(msg Message) {
	m.contentType = msg.GetContentType()
	m.header[m.contentType.string()] = msg.GetContent(m)
}

func (m *mimeBuilder) GetResultMessage(maxMsgSize uint) ([]byte, error) {
	to, ok := m.header["to"]
	if !ok {
		return nil, errors.New("wail: field 'To' doesn't provided")
	}

	date := time.Now().Format(time.RFC1123Z)

	out := fmt.Sprintf("Date:%s\r\n", date)
	out += fmt.Sprintf("Subject:%s\r\n", m.header["subject"])
	out += fmt.Sprintf("From:%s\r\n", m.header["from"])
	out += fmt.Sprintf("To:%s\r\n", to)

	if cc, ok := m.header["cc"]; ok {
		out += fmt.Sprintf("Cc:%s\r\n", cc)
	}

	if bcc, ok := m.header["bcc"]; ok {
		out += fmt.Sprintf("Bcc:%s\r\n", bcc)
	}

	out += "MIME-Version: 1.0\r\n"

	if ct, ok := m.header[m.contentType.string()]; ok {
		out += ct + "\r\n"
	}

	if maxMsgSize != 0 && uint(len(out)) > maxMsgSize {
		return nil, fmt.Errorf("wail: a max message size (%d) that the server can accept has been exceeded", maxMsgSize)
	}

	h := make([]byte, 0, len(out))

	return append(h, []byte(out)...), nil
}

func splitHeader(header string) string {
	if len(header) == 0 {
		return ""
	}

	s := strings.Fields(header)

	if len(s) == 0 {
		return header
	}

	var out string

	for i := 0; i < len(s); i++ { 
		if len(s[i]) > lineLengthLimit {
			out += strings.Join(split(s[i]), "\r\n")
		} else {
			out += s[i]
		}

		out += "\r\n"
	}

	return out[:len(out)-2]
}

func split(s string) []string {
	if len(s) == 0 {
		return nil
	}

	var out []string

	for i := 0; i < len(s); i += lineLengthLimit {
		to := i + lineLengthLimit

		if to > len(s) {
			to = len(s)
		}

		out = append(out, s[i:to])
	}

	return out
}

func base64Encode(text []byte) string {
	out := base64.StdEncoding.EncodeToString(text)

	if len(out) > lineLengthLimit {
		out = strings.Join(split(out), "\r\n")
	}

	return out
}

func qpEncode(text []byte) (string, error) {
	qp := quotedprintable.NewWriter(&bytes.Buffer{})

	out := make([]byte, len(text))
	copy(out, text)

	if _, err := qp.Write(out); err != nil {
		return "", err
	}

	if err := qp.Close(); err != nil {
		return "", err
	}

	return string(out), nil
}

func makeAddrString(addr []string) string {
	var sAddr string

	for _, v := range addr {
		if len(sAddr+v)+3 > lineLengthLimit {
			sAddr += "\r\n"
		}

		sAddr += "<" + v + ">,"
	}

	return sAddr[:len(sAddr)-1]
}
