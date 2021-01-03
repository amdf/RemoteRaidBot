package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	host     = os.Getenv("RAIDBOT_HOST")
	port     = os.Getenv("RAIDBOT_PORT")
	user     = os.Getenv("RAIDBOT_USER")
	password = os.Getenv("RAIDBOT_PASS")
	botKey   = os.Getenv("RAIDBOT_TOKEN")
	dbname   = "pnzraid"
)

var db *sqlx.DB
var bot *tgbotapi.BotAPI
var sender BotSender
var sendUpdates = make(map[int64]bool)
var userNames = make(map[User]string)

func testNotif() {
	sec := time.NewTicker(time.Second)
	for range sec.C {
		for v, bSend := range sendUpdates {
			if bSend {
				sender.SendText(v, "Notif!")
			}
		}
	}
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db, err = sqlx.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	/*
		err = db.Ping()
		if err != nil {
			panic(err)
		}
	*/
	bot, err = tgbotapi.NewBotAPI(botKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//bot.RemoveWebhook()
	go testNotif()

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		panic("bot.GetUpdatesChan()")
	}

	sender.Init(bot)

	go updateUserInfo()

	for update := range updates {
		if update.CallbackQuery != nil && nil != update.CallbackQuery.From {
			//inlinemessageid and chatinstance?
			userID := User(update.CallbackQuery.From.ID)
			var chatID *int64
			var msgID *int
			if nil != update.CallbackQuery.Message {
				if nil != update.CallbackQuery.Message.Chat {
					chatID = &update.CallbackQuery.Message.Chat.ID
				}
				msgID = &update.CallbackQuery.Message.MessageID
			}

			if !userID.IsRegistered() {
				registerUser(userID, update.CallbackQuery.From)
			}

			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			processCommand(userID, chatID, msgID, update.CallbackQuery.InlineMessageID, update.CallbackQuery.Data)

			if chatID != nil && msgID != nil {
				log.Printf("%d:%d [%s]~ %s\r\n", *chatID, *msgID, update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			} else {
				log.Printf("[%s]~ %s\r\n", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			}
		}

		if update.Message != nil && update.Message.From != nil {
			userID := User(update.Message.From.ID)
			var chatID *int64
			var msgID *int

			if nil != update.Message.Chat {
				chatID = &update.Message.Chat.ID
			}
			msgID = &update.Message.MessageID

			if !userID.IsRegistered() {
				registerUser(userID, update.Message.From)
			}

			processCommand(userID, chatID, nil, "", update.Message.Text)
			if chatID != nil && msgID != nil {
				log.Printf("%d:%d [%s] %s\r\n", *chatID, *msgID, update.Message.From.UserName, update.Message.Text)
			} else {
				log.Printf("[%s] %s\r\n", update.Message.From.UserName, update.Message.Text)
			}
		}
		if update.InlineQuery != nil {
			processInlineQuery(update.InlineQuery.ID, update.InlineQuery.Query)
		}
	}
}

func registerUser(user User, tgu *tgbotapi.User) {
	var name string
	if nil == tgu {
		name = fmt.Sprintf("user%d", user)
	} else {
		if len(tgu.UserName) > 0 {
			name = tgu.UserName
		} else {
			if (len(tgu.FirstName) > 0) || (len(tgu.LastName) > 0) {
				name = tgu.FirstName + " " + tgu.LastName
			} else {
				name = fmt.Sprintf("user%d", user)
			}
		}
	}
	fmt.Printf("New user %d known as %s\r\n", user, name)
	user.SetName(name)
}
