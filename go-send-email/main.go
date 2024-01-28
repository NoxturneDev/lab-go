package main

import (
	"bytes"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"html/template"
	"os"
)

type Config struct {
	Password string
	Receiver string
	Sender   string
	Server   string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	var body bytes.Buffer
	t, err := template.ParseFiles("./index.html")

	if err != nil {
		panic(err)
	}

	if err := t.Execute(&body, struct {
		Name string
	}{
		Name: "Galih",
	}); err != nil {
		panic(err)
	}


	conf := &Config{
		Password: os.Getenv("APP_PASSWORD"),
		Receiver: os.Getenv("RECEIVER"),
		Sender:   os.Getenv("CONFIG_SENDER_EMAIL"),
		Server:   os.Getenv("CONFIG_SMPT_SERVER"),
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "noxturne.storage2@gmail.com")
	m.SetHeader("To", conf.Receiver)
	m.SetHeader("Subject", "Hi!")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(conf.Server, 587, conf.Sender, conf.Password)

	//Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
