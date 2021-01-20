package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var requestPogoName = make(map[User]bool)
var requestPogoCode = make(map[User]bool)
var requestRaidInfo = make(map[User]bool)

func removePreviousRequests(user User) {
	delete(requestPogoName, user)
	delete(requestPogoCode, user)
	delete(requestRaidInfo, user)
}

func processCommandWithArgs(userID User, chatID *int64, msgID *int, inlineMsgID string, cmdArgs []string) {

	switch cmdArgs[0] {
	case "/start":
		removePreviousRequests(userID)

		if nil != chatID {
			menuSettings(userID, *chatID)
		}
	case "/raidstart":
		if msgID != nil {
			r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
			if err == nil {
				raid := Raid(r)
				if !raid.Started() {
					raid.Start(*msgID)
				} else {
					if nil != chatID {
						sender.SendText(*chatID, "Рейд уже стартовал!")
					}
				}
			} else {
				if nil != chatID {
					sender.SendText(*chatID, "Неправильный аргумент команды")
				}
			}
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неверное использование команды")
			}
		}
	case "/raidremove":
		if msgID != nil {
			r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
			if err == nil {
				raid := Raid(r)
				raid.Finish()

			} else {
				if nil != chatID {
					sender.SendText(*chatID, "Неправильный аргумент команды")
				}
			}
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неверное использование команды")
			}
		}
	case "/joininvite":
		r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
		if err == nil {
			raid := Raid(r)
			adminID, err2 := raid.GetAdminUserID()
			if err2 == nil {
				if adminID != userID {
					userID.Vote(raid, "invite")
				}
			}
			raid.RegisterGroupMessageForStatusUpdates(inlineMsgID)
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неправильный аргумент команды")
			}
		}

	case "/joinremote":

		r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
		if err == nil {
			raid := Raid(r)
			userID.Vote(raid, "remote")
			raid.RegisterGroupMessageForStatusUpdates(inlineMsgID)
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неправильный аргумент команды")
			}
		}

	case "/joinlive":

		r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
		if err == nil {
			raid := Raid(r)
			userID.Vote(raid, "live")
			raid.RegisterGroupMessageForStatusUpdates(inlineMsgID)
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неправильный аргумент команды")
			}
		}

	case "/joinspectator":

		r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
		if err == nil {
			raid := Raid(r)
			userID.Vote(raid, "spectator")
			raid.RegisterGroupMessageForStatusUpdates(inlineMsgID)
		} else {
			if nil != chatID {
				sender.SendText(*chatID, "Неправильный аргумент команды")
			}
		}
	}

}

func processCommand(userID User, chatID *int64, msgID *int, inlineMsgID string, cmdText string) {
	if 0 == len(cmdText) {
		return
	}
	switch cmdText {
	case "/start":
		removePreviousRequests(userID)

		if nil != chatID {
			menuSettings(userID, *chatID)
		}

	case "/newraid":
		if nil != chatID {
			sender.SendText(*chatID, "Введите информацию о рейде в свободной форме:")
			requestRaidInfo[userID] = true
		}
	case "/setname":
		if nil != chatID {
			sender.SendText(*chatID, "Введите ваше имя в Pokemon Go:")
			requestPogoName[userID] = true
		}
	case "/unreg":
		userID.Unregister()
		if nil != chatID {
			if !userID.IsRegistered() {
				sender.SendText(*chatID, "теперь вы не зарегистрированы")
			} else {
				sender.SendText(*chatID, "ошибка удаления регистрации")
			}
		}
	case "/setcode":
		if nil != chatID {
			sender.SendText(*chatID, "Введите код дружбы из Pokemon Go (12 цифр):")
			requestPogoCode[userID] = true
		}
	case "/deletecode":
		userID.DeleteCode()
		if nil != chatID {
			sender.SendText(*chatID, "Ваш код дружбы удалён.")
		}
	case "/notif on":
		userID.EnableNotifications(true)
		if nil != chatID {
			if userID.IsNotificationsEnabled() {
				sender.SendText(*chatID, "Уведомления включены. Бот будет присылать информацию о новых рейдах")
			} else {
				sender.SendText(*chatID, "Уведомления выключены. Вы не будете получать информацию о рейдах")
			}
		}
	case "/notif off":
		userID.EnableNotifications(false)
		if nil != chatID {
			if userID.IsNotificationsEnabled() {
				sender.SendText(*chatID, "Уведомления включены. Бот будет присылать информацию о новых рейдах")
			} else {
				sender.SendText(*chatID, "Уведомления выключены. Вы не будете получать информацию о рейдах")
			}
		}
	case "/r3start": //debug
		if nil != chatID {
			sender.SendText(*chatID, "Рестарт.")
		}
		bot.StopReceivingUpdates()
		cmd := exec.Command(os.Args[0], "")
		cmd.Start()
		os.Exit(0)
	default:
		if cmdText[0] == '/' { //command?
			cmdArgs := strings.Split(cmdText, " ")
			if len(cmdArgs) > 1 { //command with arguments
				processCommandWithArgs(userID, chatID, msgID, inlineMsgID, cmdArgs)
			} else {
				if nil != chatID {
					sender.SendText(*chatID, "Неизвестная команда")
				}
			}
		} else { //some text?
			if nil != chatID {
				if !processAnswer(userID, *chatID, msgID, cmdText) { //plain text - possible answer to request

					menuSettings(userID, *chatID) //if not, show menu
				}
			}
		}
	}
}

//returns true if cmdText was processed
func processAnswer(userID User, chatID int64, msgID *int, cmdText string) bool {
	if requestPogoName[userID] {
		userID.SetName(cmdText)

		if userID.IsRegistered() {
			sender.SendText(chatID, "Установлено имя "+cmdText)
		} else {
			sender.SendText(chatID, "Ошибка установки имени")
		}

		delete(requestPogoName, userID)
		return true
	}
	if requestPogoCode[userID] {
		if userID.IsRegistered() {
			if userID.SetCode(cmdText) {
				sender.SendText(chatID, "Код дружбы сохранён.")
				delete(requestPogoCode, userID)
			} else {
				sender.SendText(chatID, "Неверный формат! Нужно 12 цифр, например 1111 2222 3333 (можно без пробелов). Попробуйте ещё раз:")
			}
		} else {
			sender.SendText(chatID, "Ошибка: ввод кода для несуществующего пользователя")
			delete(requestPogoCode, userID)
		}
		return true
	}
	if requestRaidInfo[userID] {
		newraid, err := NewRaid(userID, chatID, cmdText)
		ok := (err == nil)
		if ok {
			ok = newraid.ShowConfirm()
		}
		if !ok {
			sender.SendText(chatID, "Не удалось создать рейд")
		}
		delete(requestRaidInfo, userID)
		return true
	}

	return false
}

func showLink(queryID string) {

	//article := tgbotapi.NewInlineQueryResultArticle(queryID, "Разместить", "22")

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: queryID,
		IsPersonal:    true,
		CacheTime:     0,
		//Results:           []interface{}{article},
		SwitchPMText:      "Открыть диалог с ботом",
		SwitchPMParameter: "0",
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		log.Println(err)
	}
}

func processInlineQuery(queryID string, queryText string) {
	if 0 == len(queryText) {
		showLink(queryID)
		return
	}
	r, err := strconv.ParseInt(queryText, 10, 32)
	if err == nil {
		raid := Raid(r)

		article := tgbotapi.NewInlineQueryResultArticle(queryID, "Разместить", "")
		//article.Description = "Разместить рейд"

		kb := raid.GetKeyboard()
		article.ReplyMarkup = &kb

		//msg.ParseMode = tgbotapi.ModeHTML
		article.InputMessageContent = tgbotapi.InputTextMessageContent{
			Text:      raid.GetText(),
			ParseMode: tgbotapi.ModeHTML,
		}
		inlineConf := tgbotapi.InlineConfig{
			InlineQueryID: queryID,
			IsPersonal:    true,
			CacheTime:     0,
			Results:       []interface{}{article},
		}

		if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
			log.Println(err)
		}
	}
}
