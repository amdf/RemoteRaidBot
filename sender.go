package main

import (
	"errors"
	"runtime"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	errNilBotsender = "nil BotSender"
	errNoChannel    = "no channel for chat"
)

const (
	maxUsers               = 1000
	limitAllChats          = time.Second / 30
	limitSingleChat        = time.Second
	maxQueuedMessages      = 10
	maxTotalQueuedMessages = 100
)

//BotSender sends msgs to users with telegram bot api limits
type BotSender struct {
	bot            *tgbotapi.BotAPI
	senderChannels map[int64]chan tgbotapi.Chattable
	commonChannel  chan tgbotapi.Chattable
}

//Init function
func (bs *BotSender) Init(botapi *tgbotapi.BotAPI) {
	bs.bot = botapi
	bs.senderChannels = make(map[int64]chan tgbotapi.Chattable, maxUsers)
	bs.commonChannel = make(chan tgbotapi.Chattable, maxTotalQueuedMessages)
	go bs.processAllMessages()
}

//ProcessChat creates a routine for chat
func (bs *BotSender) ProcessChat(chatID int64) {
	if nil == bs {
		panic(errNilBotsender)
	}
	_, exists := bs.senderChannels[chatID]
	if !exists {
		bs.senderChannels[chatID] = make(chan tgbotapi.Chattable, maxQueuedMessages)
		go bs.processMessagesToChat(chatID)
	}
}

//SendMessage function
func (bs BotSender) SendMessage(chatID int64, newmsg tgbotapi.Chattable) (err error) {
	ch, ok := bs.senderChannels[chatID]
	if ok {
		ch <- newmsg
	} else {
		err = errors.New(errNoChannel)
	}
	return
}

//SendText function
func (bs BotSender) SendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	bs.SendMessage(chatID, msg)
}

//EditText function
func (bs BotSender) EditText(chatID int64, msgID int, text string) {
	msg := tgbotapi.NewEditMessageText(chatID, msgID, text)

	//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	bs.SendMessage(chatID, msg)
}

func (bs BotSender) processAllMessages() {
	opened := true
	for opened {
		select {
		case msg, more := <-bs.commonChannel:
			opened = more
			bs.bot.Send(msg)
			time.Sleep(limitAllChats)
		default:
		}
		runtime.Gosched()
	}
}

func (bs BotSender) processMessagesToChat(chatID int64) {
	ch, ok := bs.senderChannels[chatID]
	if ok {
		opened := true
		for opened {
			select {
			case msg, more := <-ch:
				opened = more
				bs.commonChannel <- msg
				time.Sleep(limitSingleChat)
			default:
			}
			runtime.Gosched()
		}
	}
}
