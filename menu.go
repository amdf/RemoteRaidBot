package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func menuSettings(userID User, chatID int64) {
	var menuText = "Действия:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вкл. уведомления", "/notif on"),
			tgbotapi.NewInlineKeyboardButtonData("Выкл. уведомления", "/notif off"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("новый рейд", "/newraid"),
			tgbotapi.NewInlineKeyboardButtonData("удалиться из бота", "/unreg"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, menuText)
	msg.ReplyMarkup = keyboard
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
