// Package general defines general functions
package general

import (
	user2 "TelegramBotAI/app/user"
	"bytes"
	"database/sql"
	"encoding/json"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Base defines the base structure
type Base struct {
	MySQL *sql.DB
	Bot   *tgbapi.BotAPI
	User  *user2.User
}

// Closer defines the interface for closing files
type Closer interface {
	Close() error
}

// CloseFile closes the file and logs the error if it exists
func CloseFile(arg Closer) {
	if err := arg.Close(); err != nil {
		log.Println("CloseFile: ", err)
	}
	return
}

// VerifyUser checks if the user is admin and adds the user to the database
func VerifyUser(base *Base, request string) string {
	if base.User.IsNotAdmin() {
		return "you are not admin"
	}
	id := strings.Split(request, " ")[1]

	_, err := user2.GetUserFromDB(base.MySQL, id)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			log.Printf("AddUser:GetUserFromDB: %s", err)
			return "unknown error"
		}
	} else {
		return "user already exist"
	}

	chatid, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("AddUser:Atoi: %s", err)
		return "chatID must be integer"
	}

	usr := user2.User{ChatID: int64(chatid)}
	usr.SetUser(base.MySQL)
	return "User added successfully"
}

// ParseResponse parses the response and writes it to the variable
func ParseResponse(response *http.Response, resp interface{}) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("ParseResponse:ReadAll: %s", err)
	}
	if err = json.Unmarshal(body, resp); err != nil {
		log.Printf("ParseResponse:Unmarshal: %s", err)
	}
}

// GetResponse gets the response from the server
func GetResponse(req *bytes.Buffer, url string, content string) *http.Response {
	request, err := http.NewRequest("POST", url, req)
	if err != nil {
		log.Printf("RequestAndResponse:NewRequest: %s", err)
	}
	request.Header.Set("Content-Type", content)
	request.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("RequestAndResponse:Do: %s", err)
	}
	return response
}

// MeetRequirements checks if the query meets the requirements
func MeetRequirements(query string, minLen int) bool {
	if strings.Count(query, " ") < minLen-1 {
		return false
	}
	return true
}
