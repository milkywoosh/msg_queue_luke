package utils

import "log"

type NotifEmail struct {
	message string
}

func NewNotifEmail() *NotifEmail {
	return &NotifEmail{}
}

func (e *NotifEmail) SendEmail(message string) {
	log.Printf("send email msg: %s", message)
}
