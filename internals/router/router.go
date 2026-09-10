package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"

	"msgqueue-luke.com/v2/internals/utils"
)

type StoreMain struct {
	mutex sync.Mutex
	Key   string
	Value string
}

type Server struct {
	config     utils.Config
	httpserver *http.Server
	router     *httprouter.Router
	storeTmp   *StoreMain
}

func NewServer(
	cfg utils.Config,
) (*Server, error) {

	addr := fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port)
	newRouter := httprouter.New()

	s := &Server{
		config:     cfg,
		router:     newRouter,
		httpserver: nil,
		storeTmp:   &StoreMain{}, // init pertama di main, need mutex
	}

	srv := &http.Server{
		Handler: newRouter,
		Addr:    addr,
	}

	s.httpserver = srv
	s.AddRoute()
	s.AddData()

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

	if s.storeTmp.Key == key {
		return errors.New("key berikut sudah ada, tidak dapat duplikat")
	}

	s.storeTmp.mutex.Lock()
	s.storeTmp.Key = key
	s.storeTmp.Value = value
	defer s.storeTmp.mutex.Unlock()

	return nil
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

		dataResp["data"] = addDataParams

		utils.WriteResponse(w, http.StatusAccepted, dataResp)
	})
}

func (s *Server) Start() error {
	log.Printf("\nserver start...")
	return s.httpserver.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpserver.Shutdown(ctx)

}
