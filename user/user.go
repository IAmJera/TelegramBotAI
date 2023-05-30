// Package user defines the user methods and structure
package user

import (
	"database/sql"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

// User defines user structure
type User struct {
	ChatID    int64
	Context   []Text
	CTXLen    int
	Money     float32
	GroupMode bool
}

// Text defines context structure
type Text struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// IsNotAdmin check if the user is an admin
func (u *User) IsNotAdmin() bool {
	admin, err := strconv.Atoi(os.Getenv("ADMIN_CHATID"))
	if err != nil {
		log.Printf("isNotAdmin:Atoi: %s", err)
	}
	if u.ChatID == int64(admin) {
		return false
	}
	return true
}

// ClearContext clears context
func (u *User) ClearContext() {
	u.Context = nil
	u.CTXLen = 0
}

// NotAllowed checks if the user has access to the bot
func NotAllowed(mysql *sql.DB, msg *tgbapi.MessageConfig) bool {
	if _, err := GetUserFromDB(mysql, strconv.FormatInt(msg.ChatID, 10)); err == nil {
		return false
	}
	msg.Text = "You do not have access to this bot\n" +
		"Your ChatID: " + strconv.FormatInt(msg.ChatID, 10)
	return true
}
