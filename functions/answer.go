// Package functions defines interaction functions with openai chatgpt-3.5, DALL*E and Whisper
package functions

import (
	"TelegramBotAI/general"
	"TelegramBotAI/user"
	"bytes"
	"encoding/json"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"strings"
)

const chatAPIURL = "https://api.openai.com/v1/chat/completions"

// Request defines the structure of the request to the openai api
type Request struct {
	Model   string      `json:"model"`
	Message []user.Text `json:"messages"`
	Tokens  int         `json:"max_tokens"`
}

// Response defines the response structure from api openai
type Response struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Use      `json:"usage"`
}

// Choice defines the structure of the response array from the openai api
type Choice struct {
	Msg   user.Text `json:"message"`
	Index int       `json:"index"`
}

// Use defines the prompt info structure
type Use struct {
	Prompt     int `json:"prompt_tokens"`
	Completion int `json:"completion_tokens"`
	Total      int `json:"total_tokens"`
}

// GetCTXLen return max context length
func GetCTXLen() int {
	ctxLen, err := strconv.Atoi(os.Getenv("CONTEXT_LEN"))
	if err != nil {
		log.Panicf("CONTEXT_LEN not an integer")
	}
	return ctxLen
}

// RequestGPT checks the status of the group mode and sends a request to gpt-3.5
func RequestGPT(msg *tgbapi.MessageConfig, base *general.Base) {
	if base.User.GroupMode {
		switch strings.ToLower(strings.Split(msg.Text, " ")[0]) {
		case "бот,":
			fallthrough
		case "bot,":
			getAnswer(msg, base)
		default:
			msg.Text = ""
		}
		return
	}
	getAnswer(msg, base)
}

func getAnswer(msg *tgbapi.MessageConfig, base *general.Base) {
	if base.User.GroupMode {
		base.User.ClearContext()
	}
	base.User.Context = append(base.User.Context, user.Text{Content: msg.Text, Role: "user"})
	response := general.GetResponse(bytes.NewBuffer(request(base)), chatAPIURL, "application/json")
	defer general.CloseFile(response.Body)

	var resp Response
	general.ParseResponse(response, &resp)
	for _, choice := range resp.Choices {
		msg.Text = choice.Msg.Content
		base.User.Context = append(base.User.Context, user.Text{Content: msg.Text, Role: "assistant"})
	}

	base.User.Money += float32(resp.Usage.Total) * 0.002 / 1000
	increaseContext(base.User)
}

func request(base *general.Base) []byte {
	var jsonData = Request{
		Model:   "gpt-3.5-turbo",
		Message: base.User.Context,
		Tokens:  512,
	}

	req, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("request:Marshal: %s", err)
	}
	return req
}

func increaseContext(u *user.User) {
	u.CTXLen++
	if u.CTXLen >= GetCTXLen() {
		u.ClearContext()
	}
}
