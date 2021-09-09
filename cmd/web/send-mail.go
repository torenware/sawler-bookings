package main

import (
	"github.com/tsawler/bookings-app/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"net"
	"strconv"
	"time"
)

const MailPort = 1025

func verifyMailPort() bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", strconv.Itoa(MailPort)), timeout )
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func listenForMail() bool {
	if !verifyMailPort() {
		return false
	}
	go func() {
		for {
			msg := <-app.MailChan
			sendMail(msg)
		}
	}()
	return true
}

func sendMail(msg models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = MailPort
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println("Connect failed", err)
		return
	}
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextHTML, msg.Content)

	err = email.Send(client)
	if err != nil {
		log.Println("Send failed!", err)
	} else {
		log.Println("Mail sent")
	}
}
