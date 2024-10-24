package utils

import (
	"fmt"
	"net/smtp"

	"github.com/nicolaics/jim-carrier/config"
)

func SendEmail(to, subject, body string) error {
	from := config.Envs.CompanyEmail
	password := config.Envs.CompanyEmailPassword

	// set gmail smtp server
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// email message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	// set the gmail authentification
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// send the email
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
}
