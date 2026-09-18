package service

import (
	"fmt"
	"time"

	"msgqueue-luke.com/v2/internals/db"
	"msgqueue-luke.com/v2/internals/utils"
)

type OrderProcess struct {
	Store      *db.StoreMain
	EmailNotif *utils.NotifEmail
}

func NewOrderProcess(store *db.StoreMain, emailNotif *utils.NotifEmail) *OrderProcess {
	return &OrderProcess{
		Store: store,
	}
}

func (o *OrderProcess) ProcessData(key, newval string) error {
	err := o.Store.Update(key, newval, time.Now())
	if err != nil {
		return err
	} else {
		email := o.EmailNotif
		msg := fmt.Sprintf("send update email succes key berikut: %s", key)
		email.SendEmail(msg)

		return nil
	}

}
