// Package initial defines init functions
package initial

import (
	"TelegramBotAI/general"
	"TelegramBotAI/user"
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	tgbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"time"
)

// InitBase initializes all services and returns a structure with them
func InitBase() general.Base {
	var base general.Base
	var err error

	if base.MySQL, err = initMySQL(); err != nil {
		log.Fatal(err)
	}
	if err = prepareDB(&base); err != nil {
		log.Fatal(err)
	}

	if base.Bot, err = tgbapi.NewBotAPI(os.Getenv("TGBOT_TOKEN")); err != nil {
		log.Fatal(err)
	}

	id, err := strconv.Atoi(os.Getenv("ADMIN_CHATID"))
	if err != nil {
		log.Fatal(err)
	}
	base.User = &user.User{ChatID: int64(id)}
	general.VerifyUser(&base, "/addUser "+strconv.Itoa(id))
	return base
}

func initMySQL() (*sql.DB, error) {
	addr := os.Getenv("MYSQL_ADDRESS")
	login := os.Getenv("MYSQL_USER")
	passwd := os.Getenv("MYSQL_PASSWORD")
	auth := mysql.Config{
		User:                 login,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               "bot",
		AllowNativePasswords: true,
	}
	db, err := sql.Open("mysql", auth.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func prepareDB(base *general.Base) error {
	if exist, err := tableExist(base); err != nil {
		return err
	} else if exist {
		return nil
	}

	query := "CREATE TABLE users ( id bigint, user blob);"
	if _, err := base.MySQL.ExecContext(context.Background(), query); err != nil {
		if err.Error() != "Error 1050 (42S01): Table 'users' already exists" {
			return err
		}
	}
	return nil
}

func tableExist(base *general.Base) (bool, error) {
	if _, err := base.MySQL.Query("SELECT * FROM users;"); err != nil {
		if err.Error() == "Error 1146 (42S02): Table 'bot.users' doesn't exist" {
			return false, nil
		}
		log.Printf("tableExist: %s", err)
		return false, err
	}
	return true, nil
}
