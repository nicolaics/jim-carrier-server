package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"time"

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
		To:          []string{to},
		Subject:     subject,
		Body:        body,
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

func CreateEmailBodyOfOrder(subject, name, destination, notes, packageContent string, weight, price float64) string {
	packageContents := WrapText(packageContent, 100)

	var body string

	if subject == "New Order Arrived!" {
		body = "New order just arrived to your listing!\n\n"
	} else if subject == "Re-confirm Needed!" {
		body = "Someone just modified their order!\nPlease Re-confirm it!\n\n"
	}
	
	body += "Here are the details:\n"
	body += fmt.Sprintf("\t%-15s: %s\n", "Name", name)
	body += fmt.Sprintf("\t%-15s: %s\n", "Destination", destination)
	body += fmt.Sprintf("\t%-15s: %.1f\n", "Weight", weight)
	body += fmt.Sprintf("\t%-15s: %.1f\n", "Total Price", price)
	body += fmt.Sprintf("\t%-15s: %s\n", "Package Content", packageContents[0])

	for _, line := range packageContents[1:] {
		fmt.Printf("\t%-15s  %s\n", "", line)
	}

	if notes != "" {
		body += fmt.Sprintf("\t%-15s: %s\n", "Notes", notes)
	}

	body += "\nAttached is the image of the package!\n\n"
	body += "Confirm the order before\n"
	body += fmt.Sprintf("\t\t%s 23:59 KST (GMT +09)\n", time.Now().Local().AddDate(0, 0, 2).Format("02 JAN 2006"))
	body += "If not confirmed by then, the order will automatically be cancelled!"

	return body
}
