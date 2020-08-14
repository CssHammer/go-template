package http

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/CssHammer/go-template/healthcheck"
	"github.com/CssHammer/go-template/prometheus"
	"github.com/CssHammer/go-template/service"
)

const (
	ServiceName = "http service"

	ParamID = "id"
)

type HTTPService struct {
	server      *http.Server
	log         *zap.Logger
	service     service.Service
	healthCheck *healthcheck.Service
	middlewares *prometheus.Middlewares
}

func New(
	addr string,
	log *zap.Logger,
	service service.Service,
	healthCheck *healthcheck.Service,
	middlewares *prometheus.Middlewares) *HTTPService {

	var srv HTTPService

	srv.setupServer(addr)
	srv.log = log.Named(ServiceName)
	srv.service = service
	srv.healthCheck = healthCheck
	srv.middlewares = middlewares

	return &srv
}

func (s *HTTPService) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Info("begin run", zap.String("addr", s.server.Addr))
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("end run", zap.Error(err))
		}
		s.log.Info("end run")
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.server.Shutdown(shutdownCtx)
		if err != nil {
			s.log.Error("shutdown", zap.Error(err))
		}
	}()
}

func (s *HTTPService) setupServer(addr string) {
	server := http.Server{
		Addr: addr,
	}

	server.Handler = s.setupHandler()
	s.server = &server
}

func (s *HTTPService) setupHandler() http.Handler {
	r := mux.NewRouter()

	r.Use(s.middlewares.Timer)
	r.Use(s.middlewares.Counter)

	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.HandleFunc("/healthz", s.healthCheck.HealthHandler).Methods("GET")

	r.HandleFunc("/api/user/{id}", s.getUserHandler).Methods("GET")
	r.HandleFunc("/api/user", s.postUserHandler).Methods("POST")

	return r
}
