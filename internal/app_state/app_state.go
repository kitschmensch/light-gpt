package state

import (
	"light-gpt/internal/config"
	"light-gpt/internal/logger"
	"light-gpt/internal/model"
)

type AppState struct {
	CurrentChat []model.Message
	Logger      *logger.Logger
}

func NewAppState(cfg *config.Config) *AppState {
	logFile := "app.log"
	log, err := logger.NewLogger(cfg.LogLevel, logFile)
	if err != nil {
		panic(err)
	}
	return &AppState{
		CurrentChat: []model.Message{},
		Logger:      log,
	}
}

func (s *AppState) AddMessage(role, content, timestamp string) {
	s.CurrentChat = append(s.CurrentChat, model.Message{Role: role, Content: content, Timestamp: timestamp})
}

func (s *AppState) ClearChat() {
	s.CurrentChat = []model.Message{}
}
