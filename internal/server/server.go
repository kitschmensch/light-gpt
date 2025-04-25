package server

import (
	"net/http"

	state "light-gpt/internal/app_state"
	"light-gpt/internal/config"
	"light-gpt/internal/handler"
	"light-gpt/internal/logger"

	twilio "github.com/twilio/twilio-go"
)

type Server struct {
	cfg          *config.Config
	s            *state.AppState
	logger       *logger.Logger
	twilioClient *twilio.RestClient
}

func NewServer(cfg *config.Config, s *state.AppState, logger *logger.Logger, twilioClient *twilio.RestClient) *Server {
	return &Server{cfg: cfg, s: s, logger: logger, twilioClient: twilioClient}
}

func (srv *Server) Start() error {
	http.HandleFunc("/", handler.WebhookHandler(srv.cfg, srv.s, srv.logger, srv.twilioClient))
	return http.ListenAndServe(":"+srv.cfg.Port, nil)
}
