package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func menuSettings(userID User, chatID int64) {

	msg := tgbotapi.NewMessage(chatID, "Действия:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вкл. уведомления", "/notif on"),
			tgbotapi.NewInlineKeyboardButtonData("Выкл. уведомления", "/notif off"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("удалиться из бота", "/unreg"),
		),
	)
	sender.SendMessage(msg)
}

func showRegisterButton(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Новый пользователь?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("начать", "/reg"),
		),
	)
	sender.SendMessage(msg)
}
