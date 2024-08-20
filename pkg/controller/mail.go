package controller

import (
	// "github.com/go-mail/mail"
	// "net/http"

	// "github.com/gin-gonic/gin"
	"log"

	"gopkg.in/gomail.v2"
)

func Maill(to []string ,subject string,body string) {

	m := gomail.NewMessage()

	m.SetHeader("From", "Expense-Split <ekansh.agarwal@supersixsports.com>")

	// m.SetHeader("To", "akshay.kumar@supersixsports.com")
	m.SetHeader("To",to...)

	// m.SetAddressHeader("Cc", "oliver.doe@example.com", "Oliver")

	// m.SetHeader("Subject", "Hello!")
	m.SetHeader("Subject", subject)

	// m.SetBody("text/html", "Hello <b>Kate</b> and <i>Noah</i>!")
	m.SetBody("text/plain", body)

	// m.Attach("lolcat.jpg")

	d := gomail.NewDialer("smtp.gmail.com", 587, "ekansh.agarwal@supersixsports.com", "dwfy wvli jafh mlkn")

	// Send the email to Kate, Noah and Oliver.

	if err := d.DialAndSend(m); err != nil {

		log.Panicf("Could not send email: %v",err)

	}

}
