package main

import (
	"os"
	"os/exec"
)

func processCommand(userID User, chatID int64, msgID *int, cmdText string) {
	switch cmdText {
	case "/start":
		if userID.IsRegistered() {
			menuSettings(userID, chatID)
		} else {
			showRegisterButton(chatID)
		}
	case "/raidstart":
		if nil != msgID {
			//showRaid(userID, chatID, *msgID)
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
		if wantPogoname[userID] {
			userID.Register(cmdText)
			if userID.IsRegistered() {
				sender.SendText(chatID, "Вы успешно зарегистрировались под именем "+cmdText)
			} else {
				sender.SendText(chatID, "Ошибка регистрации")
			}
			delete(wantPogoname, userID)
		} else {
			menuSettings(userID, chatID)
		}
	}
}
