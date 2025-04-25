package main

import (
	state "light-gpt/internal/app_state"
	"light-gpt/internal/config"
	"light-gpt/internal/logger"
	"light-gpt/internal/server"
	"log"

	"github.com/joho/godotenv"
	twilio "github.com/twilio/twilio-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	logFile := "app.log"
	log, err := logger.NewLogger(cfg.LogLevel, logFile)
	if err != nil {
		log.Errorf("Error initializing logger: %s", err)
	}

	appState := state.NewAppState(cfg)

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})

	srv := server.NewServer(cfg, appState, log, twilioClient)
	log.Infof("Starting server on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Errorf("Error starting server: %s", err)
	}
}
