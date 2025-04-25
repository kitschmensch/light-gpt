package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port               string
	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioSenderNumber string
	TwilioBaseURL      string
	OllamaURL          string
	ValidPhoneNumbers  string
	Model              string
	LogLevel           string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT environment variable is not set")
	}

	return &Config{
		Port:               port,
		TwilioAccountSID:   os.Getenv("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:    os.Getenv("TWILIO_AUTH_TOKEN"),
		TwilioSenderNumber: os.Getenv("TWILIO_SENDER_NUMBER"),
		TwilioBaseURL:      os.Getenv("TWILIO_BASE_URL"),
		OllamaURL:          os.Getenv("OLLAMA_URL"),
		ValidPhoneNumbers:  os.Getenv("VALID_PHONE_NUMBERS"),
		Model:              os.Getenv("MODEL"),
		LogLevel:           os.Getenv("LOG_LEVEL"),
	}, nil
}
