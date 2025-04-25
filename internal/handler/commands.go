package handler

import (
	"fmt"
	state "light-gpt/internal/app_state"
	"light-gpt/internal/config"
	"light-gpt/internal/logger"

	twilio "github.com/twilio/twilio-go"
)

type CommandContext struct {
	Config       *config.Config
	AppState     *state.AppState
	Logger       *logger.Logger
	TwilioClient *twilio.RestClient
	PhoneNumber  string
}

type Command struct {
	Description string
	Handler     func(ctx *CommandContext)
}

var Commands map[string]Command

func init() {
	Commands = map[string]Command{
		"list": {
			Description: "Show available commands",
			Handler:     List,
		},
		"1": {
			Description: "Save and clear chat",
			Handler:     SaveAndClearChat,
		},
		"2": {
			Description: "Clear chat",
			Handler:     ClearChat,
		},
	}
}

func List(ctx *CommandContext) {
	helpMessage := "Available commands:\n"
	for cmd, command := range Commands {
		helpMessage += fmt.Sprintf("%s: %s\n", cmd, command.Description)
	}
	sendSMS(ctx.Config, ctx.TwilioClient, helpMessage, ctx.PhoneNumber)
}

func SaveAndClearChat(ctx *CommandContext) {
	filename, err := saveChatToFile(ctx.AppState)
	if err != nil {
		ctx.Logger.Errorf("Error saving chat: %v", err)
		sendSMS(ctx.Config, ctx.TwilioClient, "Error saving chat: "+err.Error(), ctx.PhoneNumber)
		return
	}

	ctx.AppState.ClearChat()
	ctx.Logger.Infof("Chat saved as %v and cleared.", filename)
	sendSMS(ctx.Config, ctx.TwilioClient, "Chat saved and cleared.", ctx.PhoneNumber)
}

func ClearChat(ctx *CommandContext) {
	ctx.AppState.ClearChat()
	ctx.Logger.Info("Chat cleared.")
	sendSMS(ctx.Config, ctx.TwilioClient, "Chat cleared.", ctx.PhoneNumber)
}
