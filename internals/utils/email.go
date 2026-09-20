package utils

import "log"

type NotifEmail struct {
	message string
}

func NewNotifEmail() *NotifEmail {
	return &NotifEmail{}
}

func (e *NotifEmail) SendEmail(message string) {
	log.Printf("send email msg key: %s", message)
}

func (e *NotifEmail) ProcessData(key, newval string) error {
	e.SendEmail(key)
	return nil
}
