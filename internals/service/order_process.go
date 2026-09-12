package service

import "msgqueue-luke.com/v2/internals/db"

type OrderProcess struct {
	Store *db.StoreMain
}

func NewOrderProcess(store *db.StoreMain) *OrderProcess {
	return &OrderProcess{
		Store: store,
	}
}

func (o *OrderProcess) ProcessData(key string) error {
	return o.Store.Update(key)
}
