package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

	sender.Init(bot)
	wantPogoname := make(map[User]bool)

	for update := range updates {

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			default:
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "ответ на callback?"))
			}
		}
		if update.Message != nil { // ignore any non-Message Updates

			log.Printf("[%s] %s\r\n", update.Message.From.UserName, update.Message.Text)

			userID := User(update.Message.From.ID)
			chatID := update.Message.Chat.ID

			sender.ProcessChat(chatID)

			switch update.Message.Text {
			case "/start":
				if userID.IsRegistered() {
					sender.SendText(chatID, "вы зарегистрированы")
				} else {
					sender.SendText(chatID, "вы не зарегистрированы")
				}
			case "/reg":
				sender.SendText(chatID, "Введите ваше имя в Pokemon Go:")
				wantPogoname[userID] = true
			case "/unreg":
				userID.Unregister()
				if !userID.IsRegistered() {
					sender.SendText(chatID, "теперь вы не зарегистрированы")
				} else {
					sender.SendText(chatID, "ошибка удаления регистрации")
				}
			case "/stop":
				sender.SendText(chatID, "стоп")
			case "/r3start":
				sender.SendText(chatID, "Рестарт.")

				bot.StopReceivingUpdates()
				cmd := exec.Command(os.Args[0], "")
				cmd.Start()
				os.Exit(0)
			default:
				if wantPogoname[userID] {
					userID.Register(update.Message.Text)
					if userID.IsRegistered() {
						sender.SendText(chatID, "Вы успешно зарегистрировались под именем "+update.Message.Text)
					} else {
						sender.SendText(chatID, "Ошибка регистрации")
					}
					delete(wantPogoname, userID)
				} else {
					sender.SendText(chatID, "Команда не распознана")
				}
			}

		}
	}
}

/*
//msg.ReplyToMessageID = update.Message.MessageID

	msg.Text = "commands:"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("настроить", "/settings"), //TODO: process callbacks
			tgbotapi.NewInlineKeyboardButtonData("перевести", "/convert"),
		),
	)
	bot.Send(msg)
*/
