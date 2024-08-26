package controller

import (
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

func Maill(to []string ,subject string,body string) {

	m := gomail.NewMessage()

	m.SetHeader("From", "Expense-Split <ekansh.agarwal@supersixsports.com>")

	m.SetHeader("To",to...)

	m.SetHeader("Subject", subject)

	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "ekansh.agarwal@supersixsports.com", os.Getenv("MAIL_PASS"))

	if err := d.DialAndSend(m); err != nil {

		log.Panicf("Could not send email: %v",err)

	}

}
