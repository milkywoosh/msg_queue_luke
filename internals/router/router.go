package router

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"msgqueue-luke.com/v2/internals/utils"
)

type Server struct {
	config     utils.Config
	httpserver *http.Server
	router     *httprouter.Router
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
	}

	srv := &http.Server{
		Handler: newRouter,
		Addr:    addr,
	}

	s.httpserver = srv
	s.AddRoute()

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

func (s *Server) Start() error {
	log.Printf("\nserver start...")
	return s.httpserver.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpserver.Shutdown(ctx)

}
