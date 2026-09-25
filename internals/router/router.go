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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rabbitmq/amqp091-go"
	"msgqueue-luke.com/v2/internals/db"
	"msgqueue-luke.com/v2/internals/mail"
	msgqueue "msgqueue-luke.com/v2/internals/msg_queue"
	"msgqueue-luke.com/v2/internals/service"
	"msgqueue-luke.com/v2/internals/storage"
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
	pub        *msgqueue.PublisherPool //PublisherChan
	mu         sync.Mutex
	email      *utils.Config
	s3Client   storage.ObjectS3
}

func NewServer(
	cfg utils.Config,
	pub *msgqueue.PublisherPool, // prev: *msgqueue.PublisherChan
	service *service.OrderProcess,
	s3Client storage.ObjectS3,
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
		s3Client:   s3Client,
	}

	srv := &http.Server{
		Handler: newRouter,
		Addr:    addr,
	}

	s.httpserver = srv
	s.AddRoute()
	s.AddData()
	s.AddDataMultipart()
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

func (s *Server) AddDataMultipart() {
	s.router.POST("/api/v1/add-data-multipart", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		dataResp := make(map[string]any)
		tokenInfo := struct {
			Bucket, SubDir, Email string
		}{
			"scmt", "PGC001", "auliya.lukman@sigma.co.id",
		}

		keyReq := r.FormValue("key")
		valueReq := r.FormValue("value")
		file1, header1, err := r.FormFile("file1")
		if err != nil {
			msg := fmt.Sprintf("err empty file 1, %s", err.Error())
			dataResp["message"] = msg
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}
		file2, header2, err := r.FormFile("file2")
		if err != nil {
			msg := fmt.Sprintf("err empty file 2, %s", err.Error())
			dataResp["message"] = msg
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		err = s.storeData(keyReq, valueReq)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		}

		msgPub := amqp091.Publishing{
			ContentType: "text/plain",
			UserId:      "lukerbtmq",
			Timestamp:   time.Now(),
			Body:        []byte(keyReq),
		}

		errCreatePub := s.pub.Publish(
			r.Context(),
			"order.exchange",
			"order.create",
			msgPub,
		)

		if errCreatePub != nil {
			log.Printf("publish test: %s", errCreatePub.Error())

			// note: prob need to send to error log info, no stoper
			// service.recordLog(identifier, errMessage)

			dataResp["message"] = errCreatePub.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		} else {
			subject := "A test email"
			content := fmt.Sprintf(`
			<h1>Hello %s</h1>
			<p>This is a test message from Lukman email queue</a></p>
			`, keyReq)
			to := []string{tokenInfo.Email}
			// attachFiles := []string{"../../note.txt", "../../readme.md"}
			attachFiles := []mail.AttachFileDetail{}
			att1 := mail.NewAttachFileDetail(file1, header1.Filename, "text/csv")
			att2 := mail.NewAttachFileDetail(file2, header2.Filename, "text/csv")
			attachFiles = append(attachFiles, att1, att2)

			newEmail := mail.NewEmailPayload(
				subject,
				content,
				to,
				nil,
				nil,
				nil, // attachFiles,
			)

			body, err := json.Marshal(newEmail)
			if err != nil {
				dataResp["message"] = err.Error()
				utils.WriteErrorResponse(
					w,
					http.StatusBadRequest,
					dataResp,
				)
			}

			msgEmailNotif := amqp091.Publishing{
				ContentType: "application/json",
				UserId:      "lukerbtmq",
				Timestamp:   time.Now(),
				Body:        []byte(body),
			}
			errNotifEmail := s.pub.Publish(
				r.Context(),
				"order.exchange",
				"order.notif.email",
				msgEmailNotif, // cuman message di-publish
			)

			if errNotifEmail != nil {
				log.Printf("notif email: %s", errNotifEmail.Error())
			}
		}

		dataResp["data"] = valueReq
		dataResp["message"] = fmt.Sprintf("order %s dalam antrian update", keyReq)

		utils.WriteResponse(w, http.StatusAccepted, dataResp)
	})
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

		msgPub := amqp091.Publishing{
			ContentType: "text/plain",
			UserId:      "lukerbtmq",
			Timestamp:   time.Now(),
			Body:        []byte(addDataParams.Key),
		}

		errCreatePub := s.pub.Publish(
			r.Context(),
			"order.exchange",
			"order.create",
			msgPub,
		)

		if errCreatePub != nil {
			log.Printf("publish test: %s", errCreatePub.Error())

			// note: prob need to send to error log info, no stoper
			// service.recordLog(identifier, errMessage)

			dataResp["message"] = errCreatePub.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
		} else {

			msgEmailNotif := amqp091.Publishing{
				ContentType: "application/json",
				UserId:      "lukerbtmq",
				Timestamp:   time.Now(),
				Body:        []byte(addDataParams.Key),
			}
			errNotifEmail := s.pub.Publish(
				r.Context(),
				"order.exchange",
				"order.notif.email",
				msgEmailNotif, // cuman message di-publish
			)

			if errNotifEmail != nil {
				log.Printf("notif email: %s", errNotifEmail.Error())
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
		// nanti dapet dari token
		tokenInfo := struct {
			Bucket, Location string
		}{
			"scmt", "PGC001",
		}

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
			msg := fmt.Sprintf("err empty file 1, %s", err.Error())
			dataResp["message"] = msg
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		defer file.Close()

		csvData, err := utils.CSVReader(file)
		if err != nil {
			msg := fmt.Sprintf("err empty file 2, %s", err.Error())
			dataResp["message"] = msg
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
			msg := fmt.Sprintf("err empty file 3, %s", err.Error())
			dataResp["message"] = msg
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		_, storageDir, err := s.s3Client.PutObject(
			ctx,
			tokenInfo.Bucket,
			tokenInfo.Location,
			fileName,
			file,
		)

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["no_trans"] = noTrans
		dataResp["message"] = "berhasil upload"
		dataResp["storage_dir"] = storageDir
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
		// nanti dapet dari token
		tokenInfo := struct {
			Bucket, Location string
		}{
			"scmt", "PGC001",
		}
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

		_, _, err = s.s3Client.PutObject(
			ctx,
			tokenInfo.Bucket,
			tokenInfo.Location,
			fileName,
			file,
		)

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
}

func (s *Server) GeneratePresignedURL() {

	s.router.POST("/api/v1/presigned-url", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		ctx := r.Context()
		tokenInfo := struct {
			Bucket, SubDir string
		}{
			"scmt", "PGC001",
		}

		dataResp := make(map[string]any)
		var reqBody PresignedParams

		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		// directory where file stored at S3
		// subdir/key-unique/csvFileName

		var optFns []func(*s3.PresignOptions) = []func(*s3.PresignOptions){
			s3.WithPresignExpires(10 * time.Minute),
		}

		preSignedHttp, storageDir, err := s.s3Client.PresignObject(
			ctx,
			tokenInfo.Bucket,
			tokenInfo.SubDir,
			uuid.NewString(),
			reqBody.ContentType,
			reqBody.CsvFileName,
			optFns...,
		)

		if err != nil {
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(w, http.StatusBadRequest, dataResp)
			return
		}

		dataResp["object_key"] = storageDir
		dataResp["upload_url"] = preSignedHttp

		utils.WriteResponse(w, http.StatusCreated, dataResp)

		/*
			note abis ini langsung call upload url dan kirim csv file, vai insomnia aja
		*/
	})

}

type StreamCSVParams struct {
	Bucket   string `json:"bucket_name"` // nama skema
	SubDir   string `json:"sub_dir"`     // location/warehouse/parent
	Key      string `json:"key"`
	FileName string `json:"file_name"`
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

		result, err := s.s3Client.GetObject(
			ctx,
			reqBody.Bucket,
			reqBody.SubDir,
			reqBody.Key,
			reqBody.FileName,
		)
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

		dataResp["data"] = len(csvData.Rows)
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
