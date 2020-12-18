package main

import (
	"database/sql"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	errCreateRaid    = "raid error"
	errRaidNotActive = "raid has not started yet"
)

//Raid - Pokemon Go Raid
type Raid int

//NewRaid makes new raid
func NewRaid(userID User, chatID int64, raidInfo string) (raid Raid, err error) {
	ok := false
	defer func() {
		if !ok {
			err = errors.New(errCreateRaid)
		}
	}()
	rs := []rune(raidInfo)
	if len(rs) > 255 {
		raidInfo = string(rs[:255])
	}

	str := `INSERT INTO raids (raid_info,chat_id,user_id) VALUES ($1, $2, $3) RETURNING raid_id`

	rows, errSQL := db.Query(str, raidInfo, chatID, userID)
	if errSQL != nil {
		return
	}
	if !rows.Next() {
		return
	}

	var lastID int
	errSQL = rows.Scan(&lastID)
	if errSQL != nil {
		return
	}

	raid = Raid(lastID)
	ok = true
	rows.Close()

	return
}

func (raid Raid) String() string {
	return fmt.Sprintf("%d", raid)
}

//GetChatID return raid admin's chat id
func (raid Raid) GetChatID() (result int64, err error) {
	rows, dberr := db.Query("SELECT chat_id FROM raids where raid_id = $1", raid)

	if dberr == nil {
		if rows.Next() {
			dberr = rows.Scan(&result)
		}
		rows.Close()
	}

	err = dberr
	return
}

//Started return true if raid is active
func (raid Raid) Started() bool {
	_, err := raid.GetMsgID()
	return (nil == err)
}

//GetMsgID return msgid in admin's chat
func (raid Raid) GetMsgID() (result int, err error) {
	var msgID sql.NullInt64
	rows, dberr := db.Query("SELECT msg_id FROM raids where raid_id = $1", raid)

	if dberr == nil {
		if rows.Next() {
			dberr = rows.Scan(&msgID)
		}
		rows.Close()
	}

	err = dberr
	if err != nil {
		return
	}

	if msgID.Valid {
		result = int(msgID.Int64)
	} else {
		err = errors.New(errRaidNotActive)
	}

	return
}

//GetRaidText return raid info
func (raid Raid) GetRaidText() (result string, err error) {
	rows, dberr := db.Query("SELECT raid_info FROM raids where raid_id = $1", raid)

	if dberr == nil {
		if rows.Next() {
			dberr = rows.Scan(&result)
		}
		rows.Close()
	}

	err = dberr
	return
}

//GetUserID return raid admin's user id
func (raid Raid) GetUserID() (result int, err error) {
	rows, dberr := db.Query("SELECT user_id FROM raids where raid_id = $1", raid)

	if dberr == nil {
		if rows.Next() {
			dberr = rows.Scan(&result)
		}
		rows.Close()
	}

	err = dberr
	return
}

//ShowConfirm shows confirmation infobox to raid admin
func (raid Raid) ShowConfirm() bool {

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "/raidstart "+raid.String()),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "/raidremove "+raid.String()),
		),
	)

	chatID, err := raid.GetChatID()
	if err != nil {
		return false
	}

	msg := tgbotapi.NewMessage(chatID, "Информация введена! Теперь нажмите кнопку для подтверждения:")
	msg.ReplyMarkup = keyboard
	sender.SendMessage(chatID, msg)

	return true
}

//Delete raid completely
func (raid Raid) Delete() {
	str := `DELETE FROM raids where raid_id = $1`
	db.Exec(str, raid)
}

//Start stores msgid of admin's control panel and makes raid active
func (raid Raid) Start(msgID int) {
	str := `UPDATE raids SET msg_id = $1 where raid_id = $2`
	db.Exec(str, msgID, raid)

	raid.ShowControls()
}

//ShowControls show control panel for raid admin
func (raid Raid) ShowControls() {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалить рейд", "/raidremove "+raid.String()),
		),
	)
	chatID, err1 := raid.GetChatID()
	msgID, err2 := raid.GetMsgID()
	raidText, err3 := raid.GetRaidText()

	if nil == err1 && nil == err2 && nil == err3 {
		msg := tgbotapi.NewEditMessageText(chatID, msgID, raidText)
		msg.ReplyMarkup = &keyboard
		sender.SendMessage(chatID, msg)
	} else {
		sender.SendText(chatID, "ошибка отображения информации о рейде")
	}
}
