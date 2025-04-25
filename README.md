# Light-GPT

Use your self-hosted AI model with your dumbphone: Light-GPT is a simple relay application that accepts SMS webhooks from [Twilio SMS API](https://www.twilio.com/en-us) and forwards them to an Ollama server. The software is intended to run on your local network, and will require port forwarding to work. Full conversation context is maintained and can be cleared by texting "1".

Project is a simple proof of concept. Suggestions and comments are welcome.

## Features

- Handles incoming webhook responses.
- Processes messages with roles, content, and timestamps.
- Configurable via environment variables.

## Prerequisites

- Go 1.16 or later
- Git
- Port forwarding set up on your local network
- A Twilio developer account, with the default callback url set to http://[your public IP address]:[some random port]

## Installation

1. **Clone the repository**:

   ```sh
   git clone https://github.com/your-username/light-gpt.git
   cd light-gpt
   ```

2. **Install dependencies:**
`go get github.com/joho/godotenv`

3. **Create a .env file:**

Create a .env file in the root directory of the project with the following content:
```
# Ollama API endpoint (default local instance)
OLLAMA_URL=http://192.168.1.62:11434

# List of valid phone numbers (10 digit numbers)
VALID_PHONE_NUMBERS=+15555555555
MODEL=mistral:7b
PORT=49553 #This is the port that light-gpt will listen for webhook responses from Sinch
LOG_LEVEL=DEBUG
LOG_FILE=light_gpt.log

# Twilio configuration
TWILIO_ACCOUNT_SID=########
TWILIO_AUTH_TOKEN=##########
TWILIO_SENDER_NUMBER=+1555555556
TWILIO_BASE_URL=https://api.twilio.com/2010-04-01
```

4. Build the binary
`go build -o light-gpt`

5. Run the application
`./light-gpt`

# License
This project is licensed under the MIT License. See the LICENSE file for details.
