package storage

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"msgqueue-luke.com/v2/internals/db"
	"msgqueue-luke.com/v2/internals/utils"
)

// global var
var GlobalObjectS3 ObjectS3

// NOTE : Hook for starting all the test, and shared only in this package
func TestMain(m *testing.M) {
	ctxBg := context.Background()

	cfg, err := utils.LoadConfig("../../")
	log.Println(cfg)
	if err != nil {
		log.Printf("err load config: %s", err.Error())
		return
	}
	// variable dipake nanti
	clientS3, err := NewClientObjectS3(ctxBg, cfg.AccessKeyS3, cfg.SecretKeyS3, cfg.AddressS3)
	if err != nil {
		log.Fatalf("err init client S3: %s", err.Error())
	}
	SetupTestClientS3(clientS3)

	os.Exit(m.Run())
}

func SetupTestClientS3(client *s3.Client) {
	log.Printf("TestMain SetupTestClientS3")

	ObjectS3Init := NewSeaweedS3(client)
	GlobalObjectS3 = ObjectS3Init

}

func TestNewSeaweedS3(t *testing.T) {

	t.Run("Get Object", func(t *testing.T) {

		// object key := "lukefile<bucket-name>/PGC001/items_prod_20.csv"
		ctxBg := context.Background()
		bucket := "lukefile"
		key := "PGC001"
		fileName := "items_prod_20.csv"

		// getOutPut
		getOutPut, err := GlobalObjectS3.GetObject(ctxBg, bucket, key, fileName)
		if err != nil {
			t.Errorf("err get output: %s", err.Error())
		}

		t.Log(getOutPut)

	})

	t.Run("Delete And Get Object, must be Error", func(t *testing.T) {

		// object key := "lukefile<bucket-name>/PGC001/items_prod_20.csv"
		ctxBg := context.Background()
		bucket := "lukefile"
		key := "PGC001"
		fileName := "items_prod_20.csv"

		_, err := GlobalObjectS3.DeleteObject(ctxBg, bucket, key, fileName)
		if err != nil {

			t.Errorf("err del output: %s", err.Error())
		}

		// getOutPut
		getOutPut, err := GlobalObjectS3.GetObject(ctxBg, bucket, key, fileName)
		if err == nil {
			t.Error("get output must be error")
		}

		t.Log(getOutPut)

	})

	t.Run("returns configured SeaweedS3 instance", func(t *testing.T) {
		cfg := aws.Config{
			Region: "us-east-1",
		}
		client := s3.NewFromConfig(cfg)

		got := NewSeaweedS3(client)

		if got == nil {
			t.Fatal("NewSeaweedS3() returned nil")
		}

		seaweed, ok := got.(*SeaweedS3)
		if !ok {
			t.Fatalf("NewSeaweedS3() returned %T, want *SeaweedS3", got)
		}

		if seaweed.Client == nil {
			t.Fatal("NewSeaweedS3() did not set Client")
		}

		if seaweed.Client != client {
			t.Fatal("NewSeaweedS3() did not preserve the input client")
		}
	})

	t.Run("accepts nil client", func(t *testing.T) {
		got := NewSeaweedS3(nil)

		if got == nil {
			t.Fatal("NewSeaweedS3(nil) returned nil")
		}

		seaweed, ok := got.(*SeaweedS3)
		if !ok {
			t.Fatalf("NewSeaweedS3(nil) returned %T, want *SeaweedS3", got)
		}

		if seaweed.Client != nil {
			t.Fatal("NewSeaweedS3(nil) must get nil")
		}
	})
}

func TestDeleteAndGetS3(t *testing.T) {
	t.Run("Delete And Get Object, must be Error", func(t *testing.T) {

		// object key := "lukefile<bucket-name>/PGC001/items_prod_20.csv"
		ctxBg := context.Background()
		bucket := "lukefile"
		key := "PGC001"
		fileName := "items_prod_20.csv"

		// getOutPut
		getOutPut, err := GlobalObjectS3.GetObject(ctxBg, bucket, key, fileName)
		if err != nil {
			t.Errorf("get output must not be error: %s", err.Error())
			t.FailNow()
		}

		defer getOutPut.Body.Close()

		csvData, err := utils.CSVReader(getOutPut.Body)
		if err != nil {
			log.Printf("err parsing io.Reader data")
			t.FailNow()
		}

		store := db.NewStoreMain()

		colMap := make(map[string]int)
		for i, cols := range csvData.Columns {

			colMap[cols] = i
			log.Printf("colMap: %d. cols: %s", colMap[cols], cols)
		}

		for _, val := range csvData.Rows {
			// qrcodeCol := colMap["qrcode"]
			ProductCodeCol := colMap["product_code"]

			err := store.SetProduct(val[ProductCodeCol])
			if err != nil {
				log.Printf("err set: %s", err.Error())
				t.FailNow()
			}
		}

		doneCheck := make(map[string]bool)

		for _, val := range csvData.Rows {
			// qrcodeCol := colMap["qrcode"]
			ProductCodeCol := colMap["product_code"]

			log.Printf(":%s", val[ProductCodeCol])

			checked := doneCheck[val[ProductCodeCol]]
			if !checked {
				log.Printf("checked: %v", checked)
				prodCount, err := store.GetProductCount(val[ProductCodeCol])
				if err != nil {
					continue
				}

				doneCheck[val[ProductCodeCol]] = true
				log.Printf("prod count: %s ;count => %d", prodCount.ProductCode, prodCount.Count)
			}

		}

		// delOutput, err := GlobalObjectS3.DeleteObject(ctxBg, bucket, key, fileName)
		// if err != nil {

		// 	t.Errorf("err del output: %s", err.Error())
		// }

		// t.Log(delOutput)

	})
}
