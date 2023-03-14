package main

import (
	"fmt"
	"log"
	"net/smtp"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

func sendStatusMail(config Config, subject string, msg string, images []string) {
	log.Printf("Mail: "+subject+"   >>> ["+msg+"] Images %v\n", len(images))
	//if !config.Mail.Enabled {
	//return
	//}

	m := gomail.NewMessage()
	m.SetHeader("From", config.Mail.From)
	m.SetHeader("To", config.Mail.From)
	m.SetHeader("Subject", subject)

	body := "<p>" + msg + "</p>"
	body += "<p><div>"
	for idx, image := range images {
		m.Embed(image)
		body += fmt.Sprintf(`<img src="cid:%v"/>`, filepath.Base(image))
		if (idx+1)%4 == 0 {
			body += "</div><div>"
		}
	}
	body += "</div></p>"
	fmt.Println(body)
	m.SetBody("text/html", body)

	d := gomail.NewPlainDialer(config.Mail.Smtp, config.Mail.Port, config.Mail.From, config.Mail.Password)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

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
