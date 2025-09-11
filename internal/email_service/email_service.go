package email_service

import (
	"fmt"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(subject, body, html, recipient_email string) {
	// Create a new message
	message := gomail.NewMessage()

	// Set email headers
	message.SetHeader("From", "no-reply@clustta.com")
	message.SetHeader("To", recipient_email)
	message.SetHeader("Subject", subject)

	// Set email body
	message.SetBody("text/plain", body)

	message.AddAlternative("text/html", html)

	// Set up the SMTP dialer
	dialer := gomail.NewDialer("smtp.zoho.com", 465, "no-reply@clustta.com", "jTkYGz75pX38")

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		fmt.Println("Error:", err)
		panic(err)
	} else {
		fmt.Println("Email sent successfully!")
	}
}
