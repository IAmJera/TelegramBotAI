// Package main defines the main functions of the program
package main

import (
	functions2 "TelegramBotAI/app/functions"
	"TelegramBotAI/app/general"
	"TelegramBotAI/app/initial"
	"TelegramBotAI/app/user"
	"encoding/json"
	"fmt"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func main() {
	base := initial.InitBase()
	u := tgbapi.NewUpdate(0)
	u.Timeout = 60

	for update := range base.Bot.GetUpdatesChan(u) {
		if update.Message != nil {
			msg := tgbapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			if user.NotAllowed(base.MySQL, &msg) {
				if _, err := base.Bot.Send(msg); err != nil {
					log.Printf("NotAllowed:Send: %s", err)
				}
				continue
			}

			base.User = user.GetUser(base.MySQL, &msg)
			if update.Message.Voice != nil {
				functions2.VoiceToText(&base, &update)
			} else {
				menu(&base, &msg)
			}
			base.User.UpdateUser(base.MySQL)
		}
	}
}

func menu(base *general.Base, msg *tgbapi.MessageConfig) {
	switch strings.Split(msg.Text, " ")[0] {
	case "/start":
		fallthrough
	case "/help":
		helpMessage(msg)
	case "/userAdd":
		if !general.MeetRequirements(msg.Text, 2) {
			msg.Text = "syntax: /userAdd [id]"
			break
		}
		msg.Text = general.VerifyUser(base, msg.Text)
	case "/spendTotal":
		msg.Text = fmt.Sprintf("You spent %f$", base.User.Money)
	case "/statistic":
		getStatistic(base, msg)
	case "/context":
		base.User.ClearContext()
		msg.Text = "Context cleared"
	case "/photoGen":
		if !general.MeetRequirements(msg.Text, 2) {
			msg.Text = "syntax: /photoGen [prompt]"
			break
		}
		functions2.SendPhoto(base, msg)
	case "/groupModeToggle":
		toggleGroupMode(base.User, msg)
	default:
		functions2.RequestGPT(msg, base)
	}
	if msg.Text != "" {
		if _, err := base.Bot.Send(msg); err != nil {
			log.Printf("menu:Send: %s", err)
		}
	}
}

func toggleGroupMode(u *user.User, msg *tgbapi.MessageConfig) {
	u.ClearContext()
	if u.GroupMode {
		u.GroupMode = false
		msg.Text = "Group Mode disabled"
		return
	}
	u.GroupMode = true
	msg.Text = "Group Mode enabled"
}

func helpMessage(msg *tgbapi.MessageConfig) {
	msg.Text = "/help - Display this message\n" +
		"/spendTotal - Outputs how much money the user spent\n" +
		"/context - Delete context manually. It usually lasts for " + strconv.Itoa(functions2.GetCTXLen()) + " queries\n" +
		"/photoGen [prompt] - Generates an image based on a prompt\n" +
		"/groupModToggle - Toggle group mode\n" +
		"/userAdd [chatID] - Give the user access to the bot (Admin)\n" +
		"/statistic - Outputs how much money each user spent (Admin)\n" +
		"If you send a voice message to the bot, it will send back a text\n" +
		"As long as you are in group mode, the bot will only respond if the request starts with \"bot,\""
}

func getStatistic(base *general.Base, msg *tgbapi.MessageConfig) {
	if base.User.IsNotAdmin() {
		msg.Text = "you are not admin"
		return
	}

	rows, err := base.MySQL.Query("SELECT user FROM users")
	if err != nil {
		log.Printf("GetStatistic:Query: %s", err)
	}
	defer general.CloseFile(rows)

	result := strings.Builder{}
	var usr user.User
	for rows.Next() {
		var blobUser []byte
		if err = rows.Scan(&blobUser); err != nil {
			log.Printf("GetStatistic:Next: %s", err)
		}

		if err = json.Unmarshal(blobUser, &usr); err != nil {
			log.Printf("GetUser:Unmarshal: %s", err)
		}

		id := strconv.Itoa(int(usr.ChatID))
		userMoney := strconv.FormatFloat(float64(usr.Money), 'f', 3, 32)
		if _, err = result.WriteString(id + ": " + userMoney + "\n"); err != nil {
			log.Printf("GetStatistic:WriteString: %s", err)
		}
	}
	msg.Text = result.String()
}
