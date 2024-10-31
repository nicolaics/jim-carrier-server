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
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	}
	
	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}
	
	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	}

	buf.WriteString(m.Body)
	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\n--%s", boundary))
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

	// email message
	message := &Message{
		To: []string{to},
		Subject: subject,
		Body: body,
		Attachments: make(map[string][]byte),
	}
	// message := []byte(fmt.Sprintf("To: %s\r\n"+
	// 	"Subject: %s\r\n"+
	// 	"\r\n"+
	// 	"%s\r\n", to, subject, body))

	if attachmentUrl != "" {
		message.AttachFile(attachmentUrl, attachedFileName)
	}

	// set the gmail authentification
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// send the email
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message.ToBytes())
}
