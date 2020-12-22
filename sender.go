package main

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	errNilBotsender = "nil BotSender"
	errNilBotAPI    = "nil botapi"
	errNoChannel    = "no channel for chat"
)

const (
	maxUsers               = 1000
	limitAllChats          = time.Second / 30
	limitSingleChat        = time.Second
	maxQueuedMessages      = 10
	maxTotalQueuedMessages = 100
)

//BotMessage contains a message to send and a channel to place info about new message
type BotMessage struct {
	msg             tgbotapi.Chattable
	callbackChannel chan<- ChatAndMessage
}

//BotSender sends msgs to users with telegram bot api limits
type BotSender struct {
	bot            *tgbotapi.BotAPI
	senderChannels map[int64]chan BotMessage
	commonChannel  chan BotMessage
}

//Init function. Argument "chats" can be nil (should populate chats later with ProcessChat).
func (bs *BotSender) Init(botapi *tgbotapi.BotAPI, chats []int64) {
	if nil == botapi {
		panic(errNilBotAPI)
	}
	if nil == bs {
		panic(errNilBotsender)
	}
	bs.bot = botapi
	bs.senderChannels = make(map[int64]chan BotMessage, maxUsers)
	bs.commonChannel = make(chan BotMessage, maxTotalQueuedMessages)
	if chats != nil {
		for _, chatID := range chats {
			bs.ProcessChat(chatID)
		}
	}
	go bs.processAllMessages()
}

//ProcessChat creates a routine for chat
func (bs *BotSender) ProcessChat(chatID int64) {
	if nil == bs {
		panic(errNilBotsender)
	}
	_, exists := bs.senderChannels[chatID]
	if !exists {
		bs.senderChannels[chatID] = make(chan BotMessage, maxQueuedMessages)
		go bs.processMessagesToChat(chatID)
	}
}

//SendMessage function
func (bs BotSender) SendMessage(chatID int64, newmsg tgbotapi.Chattable) (err error) {
	ch, ok := bs.senderChannels[chatID]
	if ok {
		msgToSend := BotMessage{msg: newmsg}
		ch <- msgToSend
	} else {
		err = errors.New(errNoChannel)
	}
	return
}

//SendMessageWithCallback function. Answer will return to callback channel
func (bs BotSender) SendMessageWithCallback(chatID int64, newmsg tgbotapi.Chattable, callback chan<- ChatAndMessage) (err error) {
	ch, ok := bs.senderChannels[chatID]
	if ok {
		msgToSend := BotMessage{msg: newmsg, callbackChannel: callback}
		ch <- msgToSend
	} else {
		err = errors.New(errNoChannel)
	}
	return
}

//SendText function
func (bs BotSender) SendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	err := bs.SendMessage(chatID, msg)
	if err != nil {
		fmt.Printf("SendText error: %s\r\n", err.Error())
	}
}

//EditText function
func (bs BotSender) EditText(chatID int64, msgID int, text string) {
	msg := tgbotapi.NewEditMessageText(chatID, msgID, text)

	err := bs.SendMessage(chatID, msg)
	if err != nil {
		fmt.Printf("EditText error: %s\r\n", err.Error())
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

func (bs BotSender) processAllMessages() {
	opened := true
	for opened {
		select {
		case botMsg, more := <-bs.commonChannel:
			opened = more
			msgSent, err := bs.bot.Send(botMsg.msg)
			if err == nil && botMsg.callbackChannel != nil {
				cm := ChatAndMessage{MsgID: msgSent.MessageID, ChatID: msgSent.Chat.ID}
				go func() { botMsg.callbackChannel <- cm }()
			}
			time.Sleep(limitAllChats)
		default:
		}
		runtime.Gosched()
	}
}
