# Light-GPT

Use your self-hosted AI model with your dumbphone: Light-GPT is a simple relay application that accepts SMS webhooks from [Sinch SMS API](https://sinch.com/apis/messaging/sms/) and forwards them to an Ollama server. The software is intended to run on your local network, and will require port forwarding to work. Full conversation context is maintained and can be cleared by texting "1".

## Features

- Handles incoming webhook responses.
- Processes messages with roles, content, and timestamps.
- Configurable via environment variables.

## Prerequisites

- Go 1.16 or later
- Git
- Port forwarding set up on your local network
- A sinch accoung, with the default callback url set to http://[your public IP address]:[some random port]

## Installation

1. **Clone the repository**:

   ```sh
   git clone https://github.com/your-username/light-gpt.git
   cd light-gpt

2. **Install dependencies:**
`go get github.com/joho/godotenv`

3. **Create a .env file:**

Create a .env file in the root directory of the project with the following content:
```
# Ollama API endpoint (default local instance)
OLLAMA_URL=http://192.168.1.123:11434

# Sinch API credentials (if needed for sending SMS, not implemented in code)
SINCH_API_KEY=key
SINCH_PLAN_ID=plan
SINCH_SENDER_NUMBER=18005551234
SINCH_BASE_URL=https://us.sms.api.sinch.com/xms/v1

# List of valid phone numbers (10 digit numbers)
VALID_PHONE_NUMBERS=yournumber
MODEL=mistral:latest
PORT=49553 #This is the port that light-gpt will listen for webhook responses from Sinch
```

4. Build the binary
`go build -o light-gpt`

5. Run the application
`./light-gpt`

# License
This project is licensed under the MIT License. See the LICENSE file for details.
