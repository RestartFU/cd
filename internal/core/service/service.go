package service

import (
	"net/http"
	"time"

	"github.com/restartfu/cd/internal/ports"
	"github.com/restartfu/cd/pkg/rmux"
)

type Service struct {
	handler ports.Handler
	router  *rmux.Router
}

func NewService(handler ports.Handler, allowFunc rmux.AllowFunc) *Service {
	s := &Service{
		handler: handler,
		router:  rmux.NewRouter(),
	}
	s.router.Allow(allowFunc)
	rmux.HandleRoute(s.router, "/deploy", handler.Deploy)
	return s
}

func (s *Service) Start(addr string) error {
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
		Handler:      s.router,
	}
	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
