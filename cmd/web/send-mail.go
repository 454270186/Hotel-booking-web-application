package main

import (
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"time"
)

func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			sendMSG(msg)
		}
	}()
}

func sendMSG(m Models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)
	if err != nil {
		errorLog.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
