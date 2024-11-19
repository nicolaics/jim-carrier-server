package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"mime/multipart"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/nicolaics/jim-carrier/config"
)

type Message struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments map[string][]byte
}

func (m *Message) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(m.CC, ",")))
	}
	
	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(m.BCC, ",")))
	}
	
	buf.WriteString("MIME-Version: 1.0\r\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	}

	buf.WriteString("\r\n")
	buf.WriteString(m.Body)

	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", k))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}

func (m *Message) AttachFile(src string, attachedFileName string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	m.Attachments[attachedFileName] = b

	return nil
}

func SendEmail(to, subject, body, attachmentUrl, attachedFileName string) error {
	from := config.Envs.CompanyEmail
	password := config.Envs.CompanyEmailPassword

	// set gmail smtp server
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	body += "\n\nCopyright 2024. Jim Carrier International."
	
	// email message
	message := &Message{
		To: []string{to},
		Subject: subject,
		Body: body,
		Attachments: make(map[string][]byte),
	}

	if attachmentUrl != "" {
		err := message.AttachFile(attachmentUrl, attachedFileName)
		if err != nil {
			return err
		}
	}

	// set the gmail authentification
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// send the email
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message.ToBytes())
}
