package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var requestPogoName = make(map[User]bool)
var requestPogoCode = make(map[User]bool)
var requestRaidInfo = make(map[User]bool)

func removePreviousRequests(user User) {
	delete(requestPogoName, user)
	delete(requestPogoCode, user)
	delete(requestRaidInfo, user)
}

func processCommandWithArgs(userID User, chatID int64, msgID *int, cmdArgs []string) {

	switch cmdArgs[0] {
	case "/raidstart":
		if msgID != nil {
			r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
			if err == nil {
				raid := Raid(r)
				if !raid.Started() {
					raid.Start(*msgID)
				} else {
					sender.SendText(chatID, "Рейд уже стартовал!")
				}
			} else {
				sender.SendText(chatID, "Неправильный аргумент команды")
			}
		} else {
			sender.SendText(chatID, "Неверное использование команды")
		}
	case "/raidremove":
		if msgID != nil {
			r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
			if err == nil {
				raid := Raid(r)
				raid.Delete()
				str := "Рейд удалён. Введите (или нажмите) /start чтобы начать заново"
				sender.EditText(chatID, *msgID, str)
			} else {
				sender.SendText(chatID, "Неправильный аргумент команды")
			}
		} else {
			sender.SendText(chatID, "Неверное использование команды")
		}
	}

}

func processCommand(userID User, chatID int64, msgID *int, cmdText string) {
	switch cmdText {
	case "/start":
		removePreviousRequests(userID)
		if userID.IsRegistered() {
			menuSettings(userID, chatID)
		} else {
			showRegisterButton(chatID)
		}
	case "/newraid":
		sender.SendText(chatID, "Введите информацию о рейде в свободной форме:")
		requestRaidInfo[userID] = true
	case "/reg":
		sender.SendText(chatID, "Введите ваше имя в Pokemon Go:")
		requestPogoName[userID] = true
	case "/unreg":
		userID.Unregister()
		if !userID.IsRegistered() {
			sender.SendText(chatID, "теперь вы не зарегистрированы")
		} else {
			sender.SendText(chatID, "ошибка удаления регистрации")
		}
	case "/setcode":
		sender.SendText(chatID, "Введите код дружбы из Pokemon Go (12 цифр):")
		requestPogoCode[userID] = true
	case "/notif on":
		userID.EnableNotifications(true)
		if userID.IsNotificationsEnabled() {
			sender.SendText(chatID, "Уведомления выключены. Вы не будете получать информацию о рейдах")
		} else {
			sender.SendText(chatID, "Уведомления включены. Бот будет присылать информацию о новых рейдах")
		}
	case "/notif off":
		userID.EnableNotifications(false)
		if userID.IsNotificationsEnabled() {
			sender.SendText(chatID, "Уведомления выключены. Вы не будете получать информацию о рейдах")
		} else {
			sender.SendText(chatID, "Уведомления включены. Бот будет присылать информацию о новых рейдах")
		}
	case "/r3start": //debug
		sender.SendText(chatID, "Рестарт.")

		bot.StopReceivingUpdates()
		cmd := exec.Command(os.Args[0], "")
		cmd.Start()
		os.Exit(0)
	default:
		if cmdText[0] == '/' { //command?
			cmdArgs := strings.Split(cmdText, " ")
			if len(cmdArgs) > 1 { //command with arguments
				processCommandWithArgs(userID, chatID, msgID, cmdArgs)
			} else {
				sender.SendText(chatID, "Неизвестная команда")
			}
		} else { //some text?
			if !processAnswer(userID, chatID, msgID, cmdText) { //plain text - possible answer to request
				menuSettings(userID, chatID) //if not, show menu
			}
		}
	}
}

//returns true if cmdText was processed
func processAnswer(userID User, chatID int64, msgID *int, cmdText string) bool {
	if requestPogoName[userID] {
		userID.Register(cmdText)
		if userID.IsRegistered() {
			sender.SendText(chatID, "Вы успешно зарегистрировались под именем "+cmdText+
				"\r\nТеперь введите код дружбы из Pokemon Go (12 цифр):")
			requestPogoCode[userID] = true
		} else {
			sender.SendText(chatID, "Ошибка регистрации")
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
