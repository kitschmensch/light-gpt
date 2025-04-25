package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	state "light-gpt/internal/app_state"
	"light-gpt/internal/config"
	"light-gpt/internal/logger"
	"light-gpt/internal/model"

	twilio "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func WebhookHandler(cfg *config.Config, appState *state.AppState, log *logger.Logger, twilioClient *twilio.RestClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		log.Info("Received webhook request")
		handleWebhook(cfg, appState, w, r, log, twilioClient)
	}
}

func handleWebhook(cfg *config.Config, s *state.AppState, w http.ResponseWriter, r *http.Request, log *logger.Logger, twilioClient *twilio.RestClient) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("Unable to parse form data: %v", err)
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	data := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 {
			data[key] = values[0]
		}
	}
	log.Infof("Received Twilio webhook data: %v", data)

	phoneNumber := data["From"]
	message := data["Body"]

	// Validate the phone number
	if !strings.Contains(cfg.ValidPhoneNumbers, phoneNumber) {
		log.Error("Received text from invalid phone number!")
		log.Errorf("Phone number: %v", phoneNumber)
		log.Errorf("Message: %v", message)
		return
	}

	// Check if the message matches a command
	if command, exists := Commands[strings.ToLower(message)]; exists {
		ctx := &CommandContext{
			Config:       cfg,
			AppState:     s,
			Logger:       log,
			TwilioClient: twilioClient,
			PhoneNumber:  phoneNumber,
		}
		command.Handler(ctx)
		return
	}

	handleDefaultMessage(cfg, s, log, twilioClient, phoneNumber, message)
}

func handleDefaultMessage(cfg *config.Config, s *state.AppState, log *logger.Logger, twilioClient *twilio.RestClient, phoneNumber, message string) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	s.AddMessage("user", message, timestamp)

	body, err := buildBody(cfg, s)
	if err != nil {
		log.Errorf("Error building request body: %v", err)
		return
	}

	// Send the chat to the AI model
	log.Infof("User message: %v", message)
	log.Info("Sending chat to AI model")
	aiResponse, err := chatCompletion(cfg, body)
	if err != nil {
		log.Errorf("Error getting AI response: %v", err)
		return
	}

	// Add the AI response to the chat history
	log.Infof("AI response: %v", aiResponse.Content)
	s.AddMessage("assistant", aiResponse.Content, aiResponse.Timestamp)

	// Send the AI response back to the user
	_, err = sendSMS(cfg, twilioClient, aiResponse.Content, phoneNumber)
	if err != nil {
		log.Errorf("Error sending SMS: %v", err)
		return
	}
}

func buildBody(cfg *config.Config, s *state.AppState) ([]byte, error) {
	messages := make([]map[string]string, len(s.CurrentChat))

	for i, msg := range s.CurrentChat {
		if i == 0 {
			messages[i] = map[string]string{
				"role":    msg.Role,
				"content": cfg.OllamaInstructions + msg.Content,
			}
			continue
		}

		messages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	body := map[string]interface{}{
		"model":    cfg.Model,
		"stream":   false,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %s", err)
	}

	return jsonBody, nil
}

func chatCompletion(cfg *config.Config, body []byte) (model.Message, error) {
	endpoint := "/api/chat"
	req, err := http.NewRequest("POST", cfg.OllamaURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return model.Message{}, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.Message{}, fmt.Errorf("error making API call: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Message{}, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Message{}, fmt.Errorf("error reading response body: %s", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(respBody, &responseMap); err != nil {
		return model.Message{}, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	messageMap, ok := responseMap["message"].(map[string]interface{})
	if !ok {
		return model.Message{}, fmt.Errorf("invalid response format")
	}

	role := "assistant"
	content, ok := messageMap["content"].(string)
	if !ok {
		return model.Message{}, fmt.Errorf("invalid content format")
	}

	timestamp, ok := responseMap["created_at"].(string)
	if !ok {
		return model.Message{}, fmt.Errorf("invalid timestamp format")
	}

	return model.Message{Role: role, Content: content, Timestamp: timestamp}, nil
}

func sendSMS(cfg *config.Config, client *twilio.RestClient, message string, toNumber string) (status int, err error) {
	if client == nil {
		return 0, fmt.Errorf("twilioClient is nil")
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(toNumber)
	params.SetFrom(cfg.TwilioSenderNumber)
	params.SetBody(message)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return 0, fmt.Errorf("error sending SMS message: %s", err)
	}

	response, _ := json.Marshal(*resp)
	fmt.Printf("Twilio Response: %s\n", string(response))

	return http.StatusCreated, nil
}

func saveChatToFile(s *state.AppState) (filename string, err error) {
	if len(s.CurrentChat) == 0 {
		return "", fmt.Errorf("no chat history to save")
	}

	firstMessage := s.CurrentChat[0]
	filename = fmt.Sprintf("%s_%s.txt", firstMessage.Timestamp[:10], strings.ReplaceAll(firstMessage.Content, " ", "_"))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("unable to create file: %s", err)
	}
	defer file.Close()

	for _, msg := range s.CurrentChat {
		line := fmt.Sprintf("%s | %s\n%s\n\n", msg.Timestamp, msg.Role, msg.Content)
		if _, err := file.WriteString(line); err != nil {
			return "", fmt.Errorf("unable to write to file: %s", err)
		}
	}
	return filename, nil
}
