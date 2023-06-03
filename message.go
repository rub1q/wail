package wail

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
)

type contentType int

const (
	TextPlain contentType = iota
	TextHtml

	multipartMix
	multipartAlt
	applOctetStream
)

var contentTypes = map[contentType]string{
	TextPlain:       "text/plain",
	TextHtml:        "text/html",
	multipartMix:    "multipart/mixed",
	multipartAlt:    "multipart/alternative",
	applOctetStream: "application/octet-stream",
}

func (c contentType) string() string {
	return contentTypes[c]
}

// Boundary is used in multipart messages
var boundary = func() string {
	h := sha256.New224()
	h.Write([]byte("6MHoYQhoRORdeWi6RzQaFKK7iGYieH"))

	out := hex.EncodeToString(h.Sum(nil))
	return out[:len(out)/2]
}()

var middleBound = "--" + boundary + "\r\n"
var endBound = "--" + boundary + "--"

type Message interface {
	// GetContent returns formatted message body text
	GetContent(mb *mimeBuilder) string

	// GetContentType returns a content type of the message
	// that is used for assembling a result message
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
func (t *TextMessage) Set(ctype contentType, text []byte) {
	t.ctype = ctype
	t.text = text
}

func (t *TextMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: %s; charset=%s\r\n", t.ctype.string(), mb.charset)
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

	a.content = make([]byte, len(content))
	copy(a.content, content)
}

func (a *Attachment) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: %s\r\n", a.GetContentType().string())
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
func (m *MultipartMixedMessage) SetText(ctype contentType, text []byte) {
	m.text.Set(ctype, text)
}

// AddAttachment adds an attachment to the message
func (m *MultipartMixedMessage) AddAttachment(attach Attachment) {
	m.attachments = append(m.attachments, attach)
}

func (m *MultipartMixedMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: %s; boundary=%s\r\n", m.GetContentType().string(), boundary)
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

// SetPlainText sets a plain part of the message with specified order (priority)
//
// Note: Anti-spam software penalizing messages with very different
// text in a multipart/alternative message
func (m *MultipartAltMessage) SetPlainText(text []byte, order int) {
	txtPlain := TextMessage{}
	txtPlain.Set(TextPlain, text)

	m.msg = append(m.msg, altMessage{text: txtPlain, order: order})
}

// SetHtmlText sets an html part of the message with specified order (priority)
//
// Note: Anti-spam software penalizing messages with very different
// text in a multipart/alternative message
func (m *MultipartAltMessage) SetHtmlText(text []byte, order int) {
	txtHtml := TextMessage{}
	txtHtml.Set(TextHtml, text)

	m.msg = append(m.msg, altMessage{text: txtHtml, order: order})
}

func (m *MultipartAltMessage) GetContent(mb *mimeBuilder) string {
	content := fmt.Sprintf("Content-Type: %s; boundary=%s\r\n", m.GetContentType().string(), boundary)
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
