package wail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime" 
	"mime/quotedprintable"
	"strings"
	"time"
)

// RFC 5322 2.2.3
const lineLenghtLimit = 76

var months = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
var days = [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

type mimeBuilder struct {
	charset     Charset
	encoding    Encoding
	encoder     mime.WordEncoder
	contentType contentType
	header      map[string]string
}

func newMimeBuilder(charset Charset, encoding Encoding) *mimeBuilder {
	mb := &mimeBuilder{
		charset:  charset,
		encoding: encoding,
		header:   make(map[string]string),
	}

	if encoding == QuotedPrintable {
		mb.encoder = mime.QEncoding
	} else {
		mb.encoder = mime.BEncoding
	}

	return mb
}

func (m *mimeBuilder) EncodeHeader(value string) string {
	if len(value) == 0 {
		return value
	}

	out := m.encoder.Encode(string(m.charset), value)

	if len(out) > lineLenghtLimit {
		split(&out)
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
	m.header[string(m.contentType)] = msg.GetContent(m)
}

func (m *mimeBuilder) GetResultMessage(maxMsgSize uint) []byte {
	h := make([]byte, 0, maxMsgSize)

	date := getCurrentDateTime()

	h = append(h, []byte("Date: "+date+"\n")...)
	h = append(h, []byte("Subject: "+m.header["subject"]+"\n")...)
	h = append(h, []byte("From: "+m.header["from"]+"\n")...)
	h = append(h, []byte("To: "+m.header["to"]+"\n")...)

	if cc, ok := m.header["cc"]; ok {
		h = append(h, []byte("Cc: "+cc+"\n")...)
	}

	if bcc, ok := m.header["bcc"]; ok {
		h = append(h, []byte("Bcc: "+bcc+"\n")...)
	}

	h = append(h, []byte("MIME-Version: 1.0\r\n")...)

	if ct, ok := m.header[string(m.contentType)]; ok {
		h = append(h, []byte(ct+"\r\n")...)
	}

	return h[0:len(h)] // cut unused bytes
}

func getCurrentDateTime() string {
	now := time.Now()
	_, offset := now.Zone()

	var sign string
	if offset >= 0 {
		sign = "+"
	} else {
		sign = "-"
	}

	zone := fmt.Sprintf("%s%.2d00", sign, offset/3600)

	date := fmt.Sprintf("%s, %d %s %d %.2d:%.2d:%.2d %s",
		days[int(now.Weekday())], now.Day(), months[int(now.Month())-1], now.Year(),
		now.Hour(), now.Minute(), now.Second(), zone)

	return date
}

func split(value *string) {
	if value == nil {
		return
	}
	
	s := strings.Fields(*value)

	if len(s) == 0 {
		return
	}

	var out string

	if len(s) == 1 {
		for i := 0; i < len(*value); i += lineLenghtLimit {
			if i+lineLenghtLimit > len(*value) {
				out += (*value)[i:len(*value)] + "\r\n"
			 } else {
				out += (*value)[i:i+lineLenghtLimit] + "\r\n"
			}
		}	
	} else if len(s) > 1 {
		for i := 0; i < len(s)-1; i++ {
			
			if len(s[i]) > lineLenghtLimit {
				split(&s[i])
			} else if len(s[i+1]) > lineLenghtLimit {
				split(&s[i+1])	
			}
			
			out += s[i] 
	
			if len(s[i])+len(s[i+1]) + 1 > lineLenghtLimit {
				out += "\r\n"
			}
	
			out += " "
			out += s[i+1]
		}
	}

	*value = out
}

// func splitBody(value *string) {
// 	var out string

// 	for i := 0; i < len(*value); i += lineLenghtLimit {
// 		if i+lineLenghtLimit > len(*value) {
// 			out += (*value)[i:len(*value)] + "\r\n"
//  		} else {
// 			out += (*value)[i:i+lineLenghtLimit] + "\r\n"
// 		}
// 	}

// 	*value = out
// }

func base64Encode(text []byte) string {
	out := base64.StdEncoding.EncodeToString(text)
	if len(out) > lineLenghtLimit {
		split(&out)
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
		if len(sAddr)+len(v)+3 > lineLenghtLimit {
			sAddr += "\r\n"
		}

		sAddr += "<" + v + ">,"
	}

	return sAddr[0 : len(sAddr)-1]
}
