package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func menuSettings(userID User, chatID int64) {
	//tgbotapi.NewInlineKeyboardButtonData("удалиться из бота", "/unreg"),
	var menuText = "Ваш код Pokemon Go: <code>" + userID.GetCode() + "</code>\r\n" +
		"Вам доступны следующие действия:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вкл. уведомления", "/notif on"),
			tgbotapi.NewInlineKeyboardButtonData("Выкл. уведомления", "/notif off"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("новый рейд", "/newraid"),
			tgbotapi.NewInlineKeyboardButtonData("изменить код", "/setcode"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, menuText)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = tgbotapi.ModeHTML
	sender.SendMessage(chatID, msg)
}

func showRegisterButton(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Новый пользователь?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("начать", "/reg"),
		),
	)
	sender.SendMessage(chatID, msg)
}
