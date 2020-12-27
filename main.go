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

	var restoreChats []int64
	restoreChats, errAllChats := GetAllChats()
	if errAllChats == nil {
		fmt.Printf("Begin process %d admins and voters chats\r\n", len(restoreChats))
		sender.Init(bot, restoreChats)
	}

	subUsers, errSubUsers := GetSubscribers()
	if errSubUsers == nil {
		fmt.Printf("Begin process additional %d subscriber chats\r\n", len(subUsers))
		for _, subUserID := range subUsers {
			sender.ProcessChat(int64(subUserID))
		}
	}

	go updateUserInfo()

	for update := range updates {

		if update.CallbackQuery != nil {
			userID := User(update.CallbackQuery.From.ID)
			chatID := update.CallbackQuery.Message.Chat.ID
			msgID := update.CallbackQuery.Message.MessageID
			sender.ProcessChat(chatID)
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			processCommand(userID, chatID, &msgID, update.CallbackQuery.Data)

			log.Printf("%d:%d [%s]~ %s\r\n", chatID, msgID, update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
		}
		if update.Message != nil {
			userID := User(update.Message.From.ID)
			chatID := update.Message.Chat.ID
			msgID := update.Message.MessageID

			sender.ProcessChat(chatID)
			processCommand(userID, chatID, nil, update.Message.Text)
			log.Printf("%d:%d [%s] %s\r\n", chatID, msgID, update.Message.From.UserName, update.Message.Text)
		}
	}
}
