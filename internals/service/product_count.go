package service

import (
	"context"
	"log"

	"msgqueue-luke.com/v2/internals/db"
	"msgqueue-luke.com/v2/internals/storage"
	"msgqueue-luke.com/v2/internals/utils"
)

type ProductProcessor interface {
	Count(string) (int, error)
	GroupingItemByProductCode(bucket, key, fileName string) error
}

type PayloadProduct struct {
	Url                   string
	Bucket, Key, FileName string
	S3Path                string
	ContentType           string
}

type Products struct {
	Store     *db.StoreMain
	storageS3 storage.ObjectS3
}

func NewProducts(store *db.StoreMain, clientS3 storage.ObjectS3) ProductProcessor {
	return &Products{
		Store:     store,
		storageS3: clientS3,
	}
}

func (p *Products) Count(key string) (int, error) {
	nodeProduct, err := p.Store.GetProductCount(key)
	if err != nil {
		return 0, err
	}
	return nodeProduct.Count, nil

}

func (p *Products) GroupingItemByProductCode(bucket, key, fileName string) error {

	// p.storageS3.GetObject(nil)
	ctxTodo := context.TODO()
	// objOutput
	objOutput, err := p.storageS3.GetObject(
		ctxTodo,
		bucket,
		key,
		fileName,
	)

	if err != nil {
		return err
	}

	defer objOutput.Body.Close()

	csvData, err := utils.CSVReader(objOutput.Body)
	if err != nil {
		log.Printf("err parsing io.Reader data")
		return err
	}

	colMap := make(map[string]int)
	for i, cols := range csvData.Columns {
		colMap[cols] = i
		log.Printf("colMap: %d. cols: %s", colMap[cols], cols)
	}

	for _, val := range csvData.Rows {
		// qrcodeCol := colMap["qrcode"]
		ProductCodeCol := colMap["product_code"]

		err := p.Store.SetProduct(val[ProductCodeCol])
		if err != nil {
			log.Printf("err set: %s", err.Error())
			return err
		}
	}
	return nil

	// return p.Store.SetProduct(key)

}
