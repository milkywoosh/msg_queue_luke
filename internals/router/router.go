package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"msgqueue-luke.com/v2/internals/db"
	msgqueue "msgqueue-luke.com/v2/internals/msg_queue"
	"msgqueue-luke.com/v2/internals/service"
	"msgqueue-luke.com/v2/internals/utils"
)

type StoreMain struct {
	mutex     sync.Mutex
	Key       string
	Value     string
	UpdatedAt time.Time
}

type Server struct {
	config     utils.Config
	httpserver *http.Server
	router     *httprouter.Router
	storeTmp   *db.StoreMain
	service    *service.OrderProcess
	pub        *msgqueue.PublisherChan
	mu         sync.Mutex
	email      *utils.Config
	s3Client   *s3.Client
}

func NewServer(
	cfg utils.Config,
	pub *msgqueue.PublisherChan,
	service *service.OrderProcess,
	s3 *s3.Client,
) (*Server, error) {

	addr := fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port)
	newRouter := httprouter.New()

	s := &Server{
		config:     cfg,
		router:     newRouter,
		httpserver: nil,
		storeTmp:   service.Store, // init pertama di main, need mutex
		service:    service,
		pub:        pub,
		s3Client:   s3,
	}

	srv := &http.Server{
		Handler: newRouter,
		Addr:    addr,
	}

	s.httpserver = srv
	s.AddRoute()
	s.AddData()
	s.GetData()
	s.UploadStream()
	s.UploadStreamAsync()
	s.GeneratePresignedURL()
	s.ProcessStream()

	return s, nil
}

func (s *Server) AddRoute() {
	s.router.GET("/api/v1/call/:param1", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		p1 := p.ByName("param1")

		dataResp := make(map[string]any)
		dataResp["data"] = p1
		dataResp["message"] = "response ok"

		utils.WriteResponse(w, http.StatusAccepted, dataResp)
	})
}

type AddDataParams struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Server) storeData(key, value string) error {

	return s.storeTmp.Create(key, value)
}

func (s *Server) AddData() {
	s.router.POST("/api/v1/add-data", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		dataResp := make(map[string]any)

		var addDataParams AddDataParams
		err := json.NewDecoder(r.Body).Decode(&addDataParams)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		}

		err = s.storeData(addDataParams.Key, addDataParams.Value)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		}

		ch := s.pub.GetCh()

		errCreatePub := msgqueue.PublishOrder(
			r.Context(),
			ch,
			"order.exchange",
			"order.create",
			addDataParams.Key, // id order : says ORD001ITEM
			"lukerbtmq",
		)
		if errCreatePub != nil {
			log.Printf("publish test: %s", errCreatePub.Error())

			// note: prob need to send to error log info, no stoper

			// service.recordLog(identifier, errMessage)

			// dataResp["message"] = err.Error()
			// utils.WriteErrorResponse(
			// 	w,
			// 	http.StatusBadRequest,
			// 	dataResp,
			// )
			// return
		}

		if errCreatePub == nil {
			errNotifEmail := msgqueue.PublishOrder(
				r.Context(),
				ch,
				"order.exchange",
				"order.notif.email",
				addDataParams.Key, // id order : says ORD001ITEM
				"lukerbtmq",
			)
			if errNotifEmail != nil {
				log.Printf("publish email: %s", errNotifEmail.Error())
			}
		}

		dataResp["data"] = addDataParams
		dataResp["message"] = fmt.Sprintf("order %s dalam antrian update", addDataParams.Key)

		utils.WriteResponse(w, http.StatusAccepted, dataResp)
	})
}

func (s *Server) GetData() {
	s.router.GET("/api/v1/get-data/:key", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		dataResp := make(map[string]any)

		key := p.ByName("key")
		data := s.storeTmp.Fetch(key)

		if data == nil {
			log.Printf("GetData data key ksoong %s", key)
			dataResp["message"] = fmt.Sprintf("GetData data key ksoong %s", key)
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		}

		dataResp["data"] = data
		utils.WriteResponse(
			w, http.StatusOK, dataResp,
		)
	})

}

type UploadParams struct {
	CsvFile *multipart.FileHeader `json:"csv_file"`
	NoTrans string                `json:"no_trans"`
}

// oke success
func (s *Server) UploadStream() {
	// csv files 100 rows

	// upload ke storage
	// background process go routine I/O process

	// response
	// langsung info response ["failed", "queue"]
	/*
		{
			"no_trans": "xxxx",
			"status_upload": ""
		}
	*/

	s.router.POST("/api/v1/upload-csv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		ctx := r.Context()
		dataResp := make(map[string]any)

		// err := r.ParseMultipartForm(10 << 20) // max 10Mb
		err := r.ParseMultipartForm(5 << 20) // max 5Mb
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		file, header, err := r.FormFile("csv_file")
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		defer file.Close()

		csvData, err := utils.CSVReader(file)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		fileName := header.Filename
		headerContentType := header.Header.Get("Content-Type")
		size := header.Size
		field := csvData.Columns
		rows := csvData.Rows

		noTrans := r.FormValue("no_trans")

		// probably the file is empty, handle if empty
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String("uploads"), // ?? uploads berdasarkan apa?
			Key:    aws.String(fmt.Sprintf("docs/%s", fileName)),
			Body:   file, // multipart.File type
			// Body:   strings.NewReader("halo dari S3 lokal"),
		})

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["no_trans"] = noTrans
		dataResp["message"] = "berhasil upload"
		dataResp["file_name"] = fileName
		dataResp["content_type"] = headerContentType
		dataResp["size"] = size
		dataResp["field"] = field
		dataResp["rows"] = rows[0:]

		utils.WriteResponse(w, http.StatusAccepted, dataResp)

	})
}

func (s *Server) UploadStreamAsync() {
	// csv files 100 rows

	// upload ke storage
	// background process go routine I/O process

	// response
	// langsung info response ["failed", "queue"]
	/*
		{
			"no_trans": "xxxx",
			"status_upload": ""
		}
	*/

	s.router.POST("/api/v1/upload-csv/async", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		ctx := r.Context()
		dataResp := make(map[string]any)

		// err := r.ParseMultipartForm(10 << 20) // max 10Mb
		err := r.ParseMultipartForm(5 << 20) // max 5Mb
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		file, header, err := r.FormFile("csv_file")
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		defer file.Close()

		csvData, err := utils.CSVReader(file)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		fileName := header.Filename
		headerContentType := header.Header.Get("Content-Type")
		size := header.Size
		field := csvData.Columns
		rows := csvData.Rows

		noTrans := r.FormValue("no_trans")

		// probably the file is empty, handle if empty
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String("uploads"), // ?? uploads berdasarkan apa?
			Key:    aws.String(fmt.Sprintf("docs/%s", fileName)),
			Body:   file, // multipart.File type
			// Body:   strings.NewReader("halo dari S3 lokal"),
		})

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["no_trans"] = noTrans
		dataResp["message"] = "berhasil upload"
		dataResp["file_name"] = fileName
		dataResp["content_type"] = headerContentType
		dataResp["size"] = size
		dataResp["field"] = field
		dataResp["rows"] = rows[0:]

		utils.WriteResponse(w, http.StatusAccepted, dataResp)

	})
}

type PresignedParams struct {
	CsvFileName string `json:"file_name"`
	ContentType string `json:"content_type"`
	BucketName  string `json:"bucket_name"`
}

func (s *Server) GeneratePresignedURL() {

	s.router.POST("/api/v1/presigned-url", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		ctx := r.Context()
		dataResp := make(map[string]any)
		var reqBody PresignedParams

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		// directory where file stored at S3
		objectKey := fmt.Sprintf(
			"%s/%s",
			uuid.NewString(),
			reqBody.CsvFileName,
		)

		// implement io.Reader

		input := &s3.PutObjectInput{
			Bucket:      aws.String(reqBody.BucketName),
			Key:         aws.String(objectKey),
			ContentType: aws.String(reqBody.ContentType),
		}

		presignClient := s3.NewPresignClient(s.s3Client)

		preSignedHttp, err := presignClient.PresignPutObject(
			ctx,
			input,
			s3.WithPresignExpires(10*time.Minute),
		)

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["object_key"] = objectKey
		dataResp["upload_url"] = preSignedHttp

		utils.WriteResponse(w, http.StatusCreated, dataResp)

		/*
			note abis ini langsung call upload url dan kirim csv file, vai insomnia aja
		*/
	})

}

type StreamCSVParams struct {
	FileName string `json:"file_name"`
	Bucket   string `json:"bucket_name"`
	Key      string `json:"key"`
}

func (s *Server) ProcessStream() {
	// proses 100 rows csv
	// req.body.no_transaction

	s.router.POST("/api/v1/process-stream-csv", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		var reqBody StreamCSVParams
		ctx := r.Context()
		dataResp := make(map[string]any)

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(reqBody.Bucket),
			Key:    aws.String(fmt.Sprintf("%s/%s", reqBody.Key, reqBody.FileName)),
		})

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		defer result.Body.Close()

		csvData, err := utils.CSVReader(result.Body)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["data"] = csvData
		dataResp["message"] = "success get data"

		utils.WriteResponse(w, http.StatusAccepted, dataResp)

		//  response
		/*
			{
				"no_trans": "xxxx",
				"status_upload" : ["failed", "queue", "success"]
				"data" 			: nil | []rows{}
			}
		*/

		/*
			download csv dengan update field
		*/

	})
}

func (s *Server) Start() error {
	log.Printf("\nserver start...")
	return s.httpserver.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpserver.Shutdown(ctx)

}
