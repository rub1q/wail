package wail

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type contentType string

const (
	textPlain contentType = "text/plain"
	textHtml  contentType = "text/html"
	multipartMix    contentType = "multipart/mixed"
	multipartAlt    contentType = "multipart/alternative"
	applOctetStream contentType = "application/octet-stream"
)

var boundary = getBoundaryValue()
var middleBound = "--" + boundary + "\r\n"
var endBound = "--" + boundary + "--"

type Message interface {
	GetContent(mb *mimeBuilder) string
	GetContentType() contentType
}

type TextMessage struct {
	ctype contentType
	text  []byte
}

// NewTextMessage creates a new text message object
func NewTextMessage() TextMessage {
	return TextMessage{}
}

// Set sets a text content type (plain or html) and message text
func (t *TextMessage) Set(ctype string, text []byte) error {
	ctype = strings.ToLower(ctype)
	
	if ctype != "plain" && ctype != "html" {
		return errors.New("wail: invalid text content type")
	}

	if ctype == "plain" {
		t.ctype = textPlain
	} else {
		t.ctype = textHtml
	}
	
	t.text = text

	return nil
}

func (t *TextMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: %s; charset=%s\r\n", t.ctype, mb.charset)
	content += fmt.Sprintf("Content-Transfer-Encoding: %s\r\n", mb.encoding)
	content += "\r\n"

	content += mb.EncodeBody(t.text)

	return content
}

func (t *TextMessage) GetContentType() contentType {
	return t.ctype
}

type Attachment struct {
	content []byte
	name    string
}

// NewAttachment creates a new attachment object
func NewAttachment() Attachment {
	return Attachment{}
}

// ReadFromFile reads the content of a file that is stored in filePath
func (a *Attachment) ReadFromFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	a.name = info.Name()

	a.content = make([]byte, len(buf))
	copy(a.content, buf)

	return nil
}

// SetAsBinary sets names and file content in cases when you can't read 
// it from file (e.g. a file content stores in DB)
func (a *Attachment) SetAsBinary(name string, content []byte) {
	a.name = name
	a.content = content
}

func (a *Attachment) GetContent(mb *mimeBuilder) string {
	content := "Content-Type: application/octet-stream\r\n"
	content += fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", a.name)
	content += fmt.Sprintf("Content-Transfer-Encoding: %s\r\n", mb.encoding)
	content += "\r\n"

	content += mb.EncodeBody(a.content)

	return content
}

func (a *Attachment) GetContentType() contentType {
	return applOctetStream
}

type MultipartMixedMessage struct {
	text        TextMessage
	attachments []Attachment
}

// NewMultipartMixedMessage creates a new multipart/mixed message object
func NewMultipartMixedMessage() MultipartMixedMessage {
	return MultipartMixedMessage{}
}

// SetText sets a text content type (plain or html) and message text
func (m *MultipartMixedMessage) SetText(ctype string, text []byte) error {
	return m.text.Set(ctype, text)
}

// AddAttachment adds an attachment to the message
func (m *MultipartMixedMessage) AddAttachment(attach Attachment) {
	m.attachments = append(m.attachments, attach)
}

func (m *MultipartMixedMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary)
	content += "\r\n"

	content += middleBound
	content += m.text.GetContent(mb)

	content += "\r\n"
	content += "\r\n"

	for _, attach := range m.attachments {
		content += middleBound
		content += attach.GetContent(mb)

		content += "\r\n"
		content += "\r\n"
	}

	content += endBound

	return content
}

func (m *MultipartMixedMessage) GetContentType() contentType {
	return multipartMix
}

type altMessage struct {
	text  TextMessage
	order int
}

type MultipartAltMessage struct {
	msg []altMessage
}

func NewMultipartAltMessage() MultipartAltMessage {
	return MultipartAltMessage{}
}

// Note: Anti-spam software penalizing messages with very different
// text in a multipart/alternative message
func (m *MultipartAltMessage) SetPlainText(text []byte, order int) {
	txtPlain := TextMessage{}
	txtPlain.Set("plain", text)

	m.msg = append(m.msg, altMessage{text: txtPlain, order: order})
}

// Note: Anti-spam software penalizing messages with very different
// text in a multipart/alternative message
func (m *MultipartAltMessage) SetHtmlText(text []byte, order int) {
	txtHtml := TextMessage{}
	txtHtml.Set("html", text)

	m.msg = append(m.msg, altMessage{text: txtHtml, order: order})
}

func (m *MultipartAltMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary)
	content += "\r\n"

	sort.SliceStable(m.msg, func(i, j int) bool {
		return m.msg[i].order < m.msg[j].order
	})

	for _, v := range m.msg {
		content += middleBound
		content += v.text.GetContent(mb)

		content += "\r\n"
		content += "\r\n"
	}

	content += endBound

	return content
}

func (m *MultipartAltMessage) GetContentType() contentType {
	return multipartAlt
}

// getBoundaryValue returns a boundary value for multipart messages
func getBoundaryValue() string {
	h := sha256.New224()
	h.Write([]byte("6MHoYQhoRORdeWi6RzQaFKK7iGYieH"))

	out := hex.EncodeToString(h.Sum(nil))
	return out[:len(out)/2]
}
