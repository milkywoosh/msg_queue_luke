package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/julienschmidt/httprouter"
	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/service"
	"msgqueue-luke.com/v2/internals/utils"
)

// butuh file csv [qrcode(unique), product_code]
func (s *Server) UploadItemProduct() {
	// store at sync.Map
	s.router.POST("/api/v1/upload-item-productcode", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// upload multipart/file ke s3, dapetin bucket/key/filename as file-s3
		// publish jobId
		// aync on consumer get file-s3, counting
		tokenUser := struct {
			Username   string
			BucketName string
			Warehouse  string
		}{
			Username:   "lukeaul01",
			BucketName: "lukefile",
			Warehouse:  "PGC001",
		}
		dataResp := make(map[string]any)

		csvFile, fileHeader, err := r.FormFile("file-upload1")
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		allowedContentType := map[string]bool{
			"text/csv":   true,
			"text/plain": true,
		}

		// check contet type deklarasi dari Client
		ctClient := fileHeader.Header.Get("Content-Type")
		if !allowedContentType[ctClient] {
			dataResp["message"] = "content type declaration harus csv atau text/plain"
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		// presigned-url -> receive URL for upload -> langsung upload dr sini
		// -> consumer async process counting per-product_code

		var optFns []func(*s3.PresignOptions) = []func(*s3.PresignOptions){
			s3.WithPresignExpires(10 * time.Minute),
		}

		preSignedURL, dir, err := s.s3Client.PresignObject(
			r.Context(),
			tokenUser.BucketName,
			tokenUser.Warehouse,
			"text/csv",
			fileHeader.Filename,
			optFns...,
		)

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		err = utils.UploadToPresignedURL(
			preSignedURL.Method,
			preSignedURL.URL,
			csvFile,
			fileHeader,
		)
		if err != nil {
			log.Printf("err check UploadToPresigned URL: %s", err.Error())
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		// s.storeData(uuidStr, )
		/*
				Url                           string
			    Bucket, Key, FileName string
			    ContentType
		*/
		payload := service.PayloadProduct{
			Url:         preSignedURL.URL,
			Bucket:      tokenUser.BucketName,
			Key:         tokenUser.Warehouse,
			FileName:    fileHeader.Filename,
			S3Path:      fmt.Sprintf("%s/%s/%s", tokenUser.BucketName, tokenUser.Warehouse, fileHeader.Filename),
			ContentType: fileHeader.Header.Get("ContentType"),
		}

		payloadJson, err := json.Marshal(payload)
		if err != nil {
			log.Printf("err check payloadJson marshal: %s", err.Error())
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		msgPub := amqp091.Publishing{
			ContentType: "text/plain",
			UserId:      "lukerbtmq",
			Timestamp:   time.Now(),
			Body:        []byte(payloadJson),
		}

		err = s.pub.Publish(
			r.Context(),
			"csv.exchange",
			"csv.product",
			msgPub,
		)
		if err != nil {
			log.Printf("err publishing CountingItemProduct(): %s", err.Error())
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["message"] = "ok"
		dataResp["csv_info"] = fileHeader.Filename
		dataResp["username"] = tokenUser.Username
		dataResp["storage_dir"] = dir
		dataResp["presigned_url"] = preSignedURL.URL

		utils.WriteResponse(w, http.StatusAccepted, dataResp)

	})
}

func (s *Server) ProductCount() {
	// store at sync.Map
	s.router.GET("/api/v1/product/count/:product_code", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// get data product_code ke sync.Map
		/* response:
		product_code: uniquecode001,
		count: int
		*/

		tokenUser := struct {
			Username string
		}{
			Username: "lukeaul01",
		}
		dataResp := make(map[string]any)

		productCode := p.ByName("product_code")

		nodeInfo, err := s.storeTmp.GetProductCount(productCode)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusConflict, dataResp)
			return
		}

		dataResp["message"] = "ok"
		dataResp["product_count"] = nodeInfo.Count
		dataResp["product_code"] = nodeInfo.ProductCode
		dataResp["username"] = tokenUser.Username

		utils.WriteResponse(w, http.StatusAccepted, dataResp)
	})
}
