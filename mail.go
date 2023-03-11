package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func sendMail(config Config, subject string, body string) {
	log.Println("Mail: " + subject + "   >>> [" + body + "]")
	if !config.Mail.Enabled {
		return
	}

	//from := "hoyle.hoyle@gmail.com"
	from := config.Mail.From

	//pass := "hfuekbfvfsohveqa"
	password := config.Mail.Password

	//to := "6012099198@txt.att.net"
	to := config.Mail.To

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.Mail.Smtp, config.Mail.Port),
		smtp.PlainAuth("", from, password, config.Mail.Smtp),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
