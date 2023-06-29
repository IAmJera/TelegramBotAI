// Package functions defines interaction functions with openai chatgpt-3.5, DALL*E and Whisper
package functions

import (
	"TelegramBotAI/app/general"
	"bytes"
	"encoding/json"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

// RequestPhoto defines the structure of the request to DALL*E
type RequestPhoto struct {
	Prompt string `json:"prompt"`
	Number int    `json:"n"`
	Size   string `json:"size"`
}

// ResponsePhoto defines the structure of the response from DALL*E
type ResponsePhoto struct {
	Data []URLs `json:"data"`
}

// URLs defines the structure of links to received photos
type URLs struct {
	URL string `json:"url"`
}

const photoAPIURL = "https://api.openai.com/v1/images/generations"

// SendPhoto sends a request to DALL*E and sends the received photos to the user
func SendPhoto(base *general.Base, msg *tgbapi.MessageConfig) {
	response := general.GetResponse(bytes.NewBuffer(requestPhoto(msg)), photoAPIURL, "application/json")
	defer general.CloseFile(response.Body)

	var resp ResponsePhoto
	general.ParseResponse(response, &resp)
	for _, choice := range resp.Data {
		photo := tgbapi.NewPhoto(msg.ChatID, tgbapi.FileURL(choice.URL))
		if _, err := base.Bot.Send(photo); err != nil {
			log.Printf("SentPhoto:Send: %s", err)
		}
	}
	base.User.Money += 0.02
	msg.Text = ""
}

func requestPhoto(msg *tgbapi.MessageConfig) []byte {
	prompt := strings.SplitAfterN(msg.Text, " ", 2)
	var jsonData = RequestPhoto{
		Prompt: prompt[1],
		Number: 1,
		Size:   "1024x1024",
	}

	req, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("request:Marshal: %s", err)
	}
	return req
}
