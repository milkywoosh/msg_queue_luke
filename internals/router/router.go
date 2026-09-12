package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rabbitmq/amqp091-go"

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
	ampqCh     *amqp091.Channel
}

func NewServer(
	cfg utils.Config,
	chPub *amqp091.Channel,
	service *service.OrderProcess,
) (*Server, error) {

	addr := fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port)
	newRouter := httprouter.New()

	s := &Server{
		config:     cfg,
		router:     newRouter,
		httpserver: nil,
		storeTmp:   service.Store, // init pertama di main, need mutex
		service:    service,
		ampqCh:     chPub,
	}

	srv := &http.Server{
		Handler: newRouter,
		Addr:    addr,
	}

	s.httpserver = srv
	s.AddRoute()
	s.AddData()
	s.GetData()

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

		err = msgqueue.PublishOrder(
			r.Context(),
			s.ampqCh,
			"order.exchange",
			"order.create",
			addDataParams.Key, // id order : says ORD001ITEM
			"lukerbtmq",
		)
		if err != nil {
			log.Printf("publish test: %s", err.Error())
			dataResp["message"] = err.Error()
			utils.WriteErrorResponse(
				w,
				http.StatusBadRequest,
				dataResp,
			)
			return
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

func (s *Server) Start() error {
	log.Printf("\nserver start...")
	return s.httpserver.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpserver.Shutdown(ctx)

}
