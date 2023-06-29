// Package user defines the structure and methods of the user
package user

import (
	"database/sql"
	"encoding/json"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

// SetUser adds a new user to the database
func (u *User) SetUser(mysql *sql.DB) {
	blobUser, err := json.Marshal(u)
	if err != nil {
		log.Printf("setUser:Marshal: %s", err)
	}

	query := "INSERT INTO `users` (`id`, `user`) VALUES (?, ?)"
	if _, err = mysql.Exec(query, int(u.ChatID), blobUser); err != nil {
		log.Printf("setUser:Query: %s", err)
	}
}

// UpdateUser overwrites the user record with actual data
func (u *User) UpdateUser(mysql *sql.DB) {
	blobUser, err := json.Marshal(u)
	if err != nil {
		log.Printf("UpdateUser:Marshal: %s", err)
	}

	id := strconv.FormatInt(u.ChatID, 10)
	if _, err = mysql.Exec("UPDATE `users` SET user = ? WHERE id = ?", blobUser, id); err != nil {
		log.Printf("UpdateUser:Exec: %s", err)
	}
}

// GetUser returns the user if it exists. Otherwise it returns an error
func GetUser(mysql *sql.DB, msg *tgbapi.MessageConfig) *User {
	var usr User
	usr.ChatID = msg.ChatID
	val, err := GetUserFromDB(mysql, strconv.FormatInt(msg.ChatID, 10))
	if err != nil {
		log.Printf("GetUser:GetUserFromDB: %s", err)
	}

	if err = json.Unmarshal(val, &usr); err != nil {
		log.Printf("GetUser:Unmarshal: %s", err)
	}
	return &usr
}

// GetUserFromDB returns the user (byte format) from the database
func GetUserFromDB(mysql *sql.DB, id string) ([]byte, error) {
	var blobUser []byte
	query := "SELECT user FROM users WHERE id = ?"
	if err := mysql.QueryRow(query, id).Scan(&blobUser); err != nil {
		return nil, err
	}
	return blobUser, nil
}
