// Package general defines general functions
package general

import (
	"TelegramBotAI/user"
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

// Base defines structure of general services
type Base struct {
	MySQL *sql.DB
	Bot   *tgbapi.BotAPI
	User  *user.User
}

// Closer is interface required for the CloseFile method
type Closer interface {
	Close() error
}

// CloseFile method call Close method of argument
func CloseFile(arg Closer) {
	if err := arg.Close(); err != nil {
		log.Println("CloseFile: ", err)
	}
	return
}

// VerifyUser adds the user to the allowed user list and returns the execution status
func VerifyUser(base *Base, request string) string {
	if base.User.IsNotAdmin() {
		return "you are not admin"
	}
	id := strings.Split(request, " ")[1]

	_, err := user.GetUserFromDB(base.MySQL, id)
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

	usr := user.User{ChatID: int64(chatid)}
	usr.SetUser(base.MySQL)
	return "User added successfully"
}

// ParseResponse reads and writes to the object the response from the server
func ParseResponse(response *http.Response, resp interface{}) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("ParseResponse:ReadAll: %s", err)
	}
	if err = json.Unmarshal(body, resp); err != nil {
		log.Printf("ParseResponse:Unmarshal: %s", err)
	}
}

// GetResponse sends a request to the server and returns a response
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
