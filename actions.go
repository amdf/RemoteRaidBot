package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func processCommand(userID User, chatID int64, msgID *int, cmdText string) {
	switch cmdText {
	case "/start":
		if userID.IsRegistered() {
			menuSettings(userID, chatID)
		} else {
			showRegisterButton(chatID)
		}
	case "/newraid":
		newraid, err := NewRaid(userID, chatID, "(здесь нужно будет ввести инфо)")
		ok := (err == nil)
		if ok {
			ok = newraid.ShowStart()
		}
		if !ok {
			sender.SendText(chatID, "Не удалось создать рейд")
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
	case "/r3start":
		sender.SendText(chatID, "Рестарт.")

		bot.StopReceivingUpdates()
		cmd := exec.Command(os.Args[0], "")
		cmd.Start()
		os.Exit(0)
	default:
		if cmdText[0] == '/' { //command?
			cmdArgs := strings.Split(cmdText, " ")
			if len(cmdArgs) > 1 { //command with arguments
				switch cmdArgs[0] {
				case "/raidstart":
					if msgID != nil {
						r, err := strconv.ParseInt(cmdArgs[1], 10, 32)
						if err == nil {
							//correct raid number
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
				}
			}
		} else { //some text?
			if wantPogoname[userID] { //answer to pokemon username request?
				userID.Register(cmdText)
				if userID.IsRegistered() {
					sender.SendText(chatID, "Вы успешно зарегистрировались под именем "+cmdText)
				} else {
					sender.SendText(chatID, "Ошибка регистрации")
				}
				delete(wantPogoname, userID)
			} else { //something else?
				menuSettings(userID, chatID)
			}
		}
	}
}
