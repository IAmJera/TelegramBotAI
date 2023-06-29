// Package functions defines interaction functions with openai chatgpt-3.5, DALL*E and Whisper
package functions

import (
	"TelegramBotAI/app/general"
	"bytes"
	"encoding/json"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

// Texts defines the structure of Whisper's response
type Texts struct {
	TXT string `json:"text"`
}

const voiceAPIURL = "https://api.openai.com/v1/audio/transcriptions"

// VoiceToText sends a voice message and receives text from it
func VoiceToText(base *general.Base, update *tgbapi.Update) {
	id := strconv.FormatInt(base.User.ChatID, 10)
	msg := tgbapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	getVoice(base, update)
	if err := exec.Command("ffmpeg", "-i", id+"audio.ogg", id+"audio.mp3", "-y").Run(); err != nil {
		log.Printf("VoiceToText:Command: %s", err)
	}
	composeRequest(base, writer)

	msg.Text = getResponse(body, writer)
	base.User.Money += float32(update.Message.Voice.Duration) * 0.006 / 60
	clean(id+"audio.ogg", id+"audio.mp3")
	if _, err := base.Bot.Send(msg); err != nil {
		log.Printf("VoiceToText:Send: %s", err)
	}
}

func getResponse(body *bytes.Buffer, writer *multipart.Writer) string {
	response := general.GetResponse(body, voiceAPIURL, writer.FormDataContentType())
	defer general.CloseFile(response.Body)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("getText:ReadAll: %s", err)
	}

	var res Texts
	if err = json.Unmarshal(responseBody, &res); err != nil {
		log.Printf("getText:Unmarshal: %s", err)
	}
	return res.TXT
}

func composeRequest(base *general.Base, writer *multipart.Writer) {
	id := strconv.FormatInt(base.User.ChatID, 10)
	mp3File, err := os.Open(id + "audio.mp3")
	if err != nil {
		log.Printf("ComposeRequest:Open: %s", err)
	}
	defer general.CloseFile(mp3File)

	part, err := writer.CreateFormFile("file", id+"audio.mp3")
	if err != nil {
		log.Printf("composeRequest:CreateFormFile: %s", err)
	}
	if _, err = io.Copy(part, mp3File); err != nil {
		log.Printf("composeRequest:Copy: %s", err)
	}

	if err = writer.WriteField("model", "whisper-1"); err != nil {
		log.Printf("composeRequest:WriteField: %s", err)
	}
	general.CloseFile(writer)
}

func getVoice(base *general.Base, update *tgbapi.Update) {
	voiceFileURL, err := base.Bot.GetFileDirectURL(update.Message.Voice.FileID)
	if err != nil {
		log.Printf("getVoice:GetFileDirectURL: %s", err)
	}

	resp, err := http.Get(voiceFileURL)
	if err != nil {
		log.Printf("getVoice:Get: %s", err)
	}
	defer general.CloseFile(resp.Body)

	oggFile, err := os.Create(strconv.FormatInt(base.User.ChatID, 10) + "audio.ogg")
	if err != nil {
		log.Printf("getVoice:Create: %s", err)
	}
	defer general.CloseFile(oggFile)

	if _, err = io.Copy(oggFile, resp.Body); err != nil {
		log.Printf("getVoice:Copy: %s", err)
	}
}

func clean(args ...string) {
	for _, file := range args {
		if err := os.Remove(file); err != nil {
			log.Printf("clean:Remove: %s", err)
		}
	}
}
