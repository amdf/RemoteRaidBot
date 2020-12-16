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
var wantPogoname = make(map[User]bool)

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

	sender.Init(bot)

	for update := range updates {

		if update.CallbackQuery != nil {
			log.Printf("[%s]~ %s\r\n", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)

			userID := User(update.CallbackQuery.From.ID)
			chatID := update.CallbackQuery.Message.Chat.ID

			sender.ProcessChat(chatID)
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			processCommand(userID, chatID, update.CallbackQuery.Data)
		}
		if update.Message != nil {

			log.Printf("[%s] %s\r\n", update.Message.From.UserName, update.Message.Text)

			userID := User(update.Message.From.ID)
			chatID := update.Message.Chat.ID

			sender.ProcessChat(chatID)
			processCommand(userID, chatID, update.Message.Text)

		}
	}
}
