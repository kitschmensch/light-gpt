package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"created_at"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	//Handles incoming webhook responses. URL is the set in the Sinch dashboard.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Resource created")
		handleWebhook(r)
	})

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func isValidPhoneNumber(phoneNumber string) bool {
	validPhoneNumbers := os.Getenv("VALID_PHONE_NUMBERS")
	if validPhoneNumbers == "" {
		log.Fatal("VALID_PHONE_NUMBERS environment variable is not set")
	}

	phoneNumbers := strings.Split(validPhoneNumbers, ",")
	for _, validPhoneNumber := range phoneNumbers {
		if phoneNumber == validPhoneNumber {
			return true
		}
	}
	return false
}

func handleWebhook(r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Unable to read body")
		return
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Unable to parse JSON")
		return
	}
	phoneNumber, ok := data["from"].(string)
	if !ok || !isValidPhoneNumber(phoneNumber) {
		fmt.Println("Invalid phone number")
		return
	}

	message, ok := data["body"].(string)
	if !ok {
		fmt.Println("Invalid message")
		return
	}

	//Command to clear the chat.
	if message == "1" {
		//Todo: save the chat to a file
		currentChat = []Message{}
		fmt.Println("Chat cleared")
		sendSMS("Chat cleared.", phoneNumber)
		return
	}

	timestamp, ok := data["received_at"].(string)
	if !ok {
		fmt.Println("Invalid timestamp")
		return
	}

	addMessage("user", message, timestamp)
	body, err = buildBody()
	if err != nil {
		fmt.Println("Error building request body")
		return
	}
	aiResponse, err := chatCompletion(body)
	if err != nil {
		fmt.Println("Error getting AI response")
		return
	}
	addMessage("assistant", aiResponse.Content, aiResponse.Timestamp)
	sendSMS(aiResponse.Content+"\n\n[1] Clear Chat", phoneNumber)
}

var currentChat = []Message{}

func addMessage(role string, content string, timestamp string) {
	currentChat = append(currentChat, Message{role, content, timestamp})
}

func buildBody() ([]byte, error) {
	model := os.Getenv("MODEL")
	if model == "" {
		return nil, fmt.Errorf("MODEL environment variable is not set")
	}
	messages := make([]map[string]string, len(currentChat))

	for i, msg := range currentChat {
		messages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	body := map[string]interface{}{
		"model":    model,
		"stream":   false,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %s", err)
	}

	return jsonBody, nil
}

func chatCompletion(body []byte) (Message, error) {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		fmt.Println("OLLAMA_URL environment variable is not set")
		return Message{}, fmt.Errorf("OLLAMA_URL environment variable is not set")
	}
	endpoint := "/api/chat"
	req, err := http.NewRequest("POST", ollamaURL+endpoint, bytes.NewReader(body))
	if err != nil {
		fmt.Println("error creating request")
		return Message{}, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error making API call")
		return Message{}, fmt.Errorf("error making API call: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("received non-200 response")
		return Message{}, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body")
		return Message{}, fmt.Errorf("error reading response body: %s", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(respBody, &responseMap); err != nil {
		fmt.Println("error unmarshalling response body")
		return Message{}, fmt.Errorf("error unmarshalling response body: %s", err)
	}

	messageMap, ok := responseMap["message"].(map[string]interface{})
	if !ok {
		fmt.Println("invalid response format")
		return Message{}, fmt.Errorf("invalid response format")
	}

	role := "assistant"
	content, ok := messageMap["content"].(string)
	if !ok {
		fmt.Println("invalid content format")
		return Message{}, fmt.Errorf("invalid content format")
	}
	timestamp, ok := responseMap["created_at"].(string)
	if !ok {
		fmt.Println("invalid timestamp format")
		return Message{}, fmt.Errorf("invalid timestamp format")
	}

	return Message{Role: role, Content: content, Timestamp: timestamp}, nil
}

func sendSMS(message string, to_number string) (int, error) {
	sinchURL := os.Getenv("SINCH_BASE_URL")
	if sinchURL == "" {
		fmt.Println("SINCH_BASE_URL environment variable is not set")
		return 0, fmt.Errorf("sinch_base_url environment variable is not set")
	}
	sinchPlanId := os.Getenv("SINCH_PLAN_ID")
	if sinchPlanId == "" {
		fmt.Println("SINCH_PLAN_ID environment variable is not set")
		return 0, fmt.Errorf("sinch_plan_id environment variable is not set")
	}
	endpoint := sinchURL + "/" + sinchPlanId + "/batches"

	senderNumber := os.Getenv("SINCH_SENDER_NUMBER")

	requestBody := map[string]interface{}{
		"from": senderNumber,
		"to":   []string{to_number},
		"body": message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("error marshalling json")
		return 0, fmt.Errorf("error marshalling json: %s", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("error creating request")
		return 0, fmt.Errorf("error creating request: %s", err)
	}

	//Set Headers
	req.Header.Set("Content-Type", "application/json")
	apiKey := os.Getenv("SINCH_API_KEY")
	if apiKey == "" {
		fmt.Println("SINCH_API_KEY environment variable is not set")
		return 0, fmt.Errorf("sinch_api_key environment variable is not set")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error making API call")
		return 0, fmt.Errorf("error making api call: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		fmt.Println("recieved non-201 response: ", resp.Body)
		return 0, fmt.Errorf("recieved non-201 response: %d", resp.Body)
	}

	return 201, nil
}
