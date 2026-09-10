package service

import "msgqueue-luke.com/v2/internals/db"

type OrderProcess struct {
	store *db.StoreMain
}

func NewOrderProcess(store *db.StoreMain) *OrderProcess {
	return &OrderProcess{
		&db.StoreMain{},
	}
}

func (o *OrderProcess) ProcessData(key string) error {
	return o.store.Update(key)
}
