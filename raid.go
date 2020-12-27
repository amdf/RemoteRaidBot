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

//ChatAndMessage struct
type ChatAndMessage struct {
	ChatID int64 `db:"chat_id"`
	MsgID  int   `db:"msg_id"`
}

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

//GetAdminUserID function
func (raid Raid) GetAdminUserID() (user User, err error) {
	rows, dberr := db.Query("SELECT user_id FROM raids where raid_id = $1", raid)

	if dberr == nil {
		if rows.Next() {
			dberr = rows.Scan(&user)
		}
		rows.Close()
	}

	err = dberr
	return
}

//GetAdminChatID return raid admin's chat id
func (raid Raid) GetAdminChatID() (result int64, err error) {
	adminUserID, dberr := raid.GetAdminUserID()
	if dberr == nil {
		result = int64(adminUserID)
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

//GetRaidInfo return raid info
func (raid Raid) GetRaidInfo() (result string, err error) {
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

	chatID, err := raid.GetAdminChatID()
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
	str = `DELETE FROM votes where raid_id = $1`
	db.Exec(str, raid)
	str = `DELETE FROM chats where raid_id = $1`
	db.Exec(str, raid)
}

//Stop - delete raid and remove vote buttons in all chats, etc.
func (raid Raid) Stop() {
	str := "Рейд удалён. Введите (или нажмите) /start чтобы начать заново"

	chatID, err1 := raid.GetAdminChatID()
	msgID, err2 := raid.GetMsgID()
	if nil == err1 && nil == err2 {
		sender.EditText(chatID, msgID, str)
	}

	str = "Рейд закончен"

	var cm []ChatAndMessage
	err := db.Select(&cm, "SELECT chat_id, msg_id FROM chats WHERE raid_id = $1", raid)
	if err == nil {
		for _, voter := range cm {
			sender.EditText(voter.ChatID, voter.MsgID, str)
		}
	}
	raid.Delete()
}

//Start stores msgid of admin's control panel and makes raid active
func (raid Raid) Start(msgID int) {
	str := `UPDATE raids SET msg_id = $1 where raid_id = $2`
	db.Exec(str, msgID, raid)

	raid.UpdateRaidAdminInfo() //for admin

	subUsers, err := GetSubscribers()
	fmt.Printf("Send raid announce to %d chats:\r\n", len(subUsers))
	if err == nil {
		adminUserID, _ := raid.GetAdminUserID()
		for _, subscriberUserID := range subUsers {
			if subscriberUserID != adminUserID {
				fmt.Printf("Send announce to %d\r\n", subscriberUserID)
				raid.CreateUserInfo(subscriberUserID)
			}
		}
	}
}

//GetKeyboard func
func (raid Raid) GetKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пригласите", "/joininvite "+raid.String()),
			tgbotapi.NewInlineKeyboardButtonData("Достаю", "/joinremote "+raid.String()),
			tgbotapi.NewInlineKeyboardButtonData("Вживую", "/joinlive "+raid.String()),
		),
	)
}

//GetAdminKeyboard func
func (raid Raid) GetAdminKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вживую", "/joinlive "+raid.String()),
			tgbotapi.NewInlineKeyboardButtonData("Достаю", "/joinremote "+raid.String()),
			tgbotapi.NewInlineKeyboardButtonData("Удалить", "/raidremove "+raid.String()),
		),
	)
}

//UpdateRaidAdminInfo show control panel for raid admin
func (raid Raid) UpdateRaidAdminInfo() {
	keyboard := raid.GetAdminKeyboard()
	chatID, err1 := raid.GetAdminChatID()
	msgID, err2 := raid.GetMsgID()

	raidText := raid.GetText()

	if nil == err1 && nil == err2 {
		msg := tgbotapi.NewEditMessageText(chatID, msgID, raidText)
		msg.ReplyMarkup = &keyboard
		msg.ParseMode = tgbotapi.ModeHTML
		sender.SendMessage(chatID, msg)
	}
}

//RegisterChatForStatusUpdates function
func (raid Raid) RegisterChatForStatusUpdates(chatID int64, msgID int) {
	str := `INSERT INTO chats (raid_id, chat_id, msg_id) VALUES ($1, $2, $3)`
	db.Exec(str, raid, chatID, msgID)
}

//RaidPlayer type
type RaidPlayer struct {
	Role string         `db:"raid_role"`
	Name string         `db:"pogoname"`
	Code sql.NullString `db:"pogocode"`
}

//GetText returns complete raid info text
func (raid Raid) GetText() (raidText string) {
	raidInfo, err := raid.GetRaidInfo()
	if err == nil {
		raidText += "Рейд создан в @pnzraidbot\r\n" + raidInfo

		var raidPlayers []RaidPlayer
		db.Select(&raidPlayers, `SELECT raid_role,pogoname, pogocode FROM votes FULL OUTER JOIN players USING (user_id) WHERE raid_id = $1`, raid)

		if len(raidPlayers) > 0 {
			raidText += "\r\nПригласите:"
			for _, playerInfo := range raidPlayers {
				if playerInfo.Role == "invite" {
					code := " (нет кода)"
					if playerInfo.Code.Valid {
						code = "   Код: <code>" + playerInfo.Code.String + "</code>"
					}
					raidText += "\r\n • <b>" + playerInfo.Name + "</b>" + code
				}
			}
			raidText += "\r\nДостаю:"
			for _, playerInfo := range raidPlayers {
				if playerInfo.Role == "remote" {
					raidText += "\r\n • <b>" + playerInfo.Name + "</b>"
				}
			}
			raidText += "\r\nВживую:"
			for _, playerInfo := range raidPlayers {
				if playerInfo.Role == "live" {
					raidText += "\r\n • <b>" + playerInfo.Name + "</b>"
				}
			}
		}

	} else {
		raidText = "ошибка"
	}
	return
}

//CreateUserInfo - show raid info for user to obtain their vote
func (raid Raid) CreateUserInfo(user User) {
	keyboard := raid.GetKeyboard()
	raidText := raid.GetText()

	chatID := int64(user)

	msg := tgbotapi.NewMessage(chatID, raidText)
	msg.ReplyMarkup = &keyboard
	msg.ParseMode = tgbotapi.ModeHTML

	go func() {
		callback := make(chan ChatAndMessage)
		sender.SendMessageWithCallback(chatID, msg, callback)
		msgInfo := <-callback //wait for actual sending
		sender.ProcessChat(msgInfo.ChatID)

		raid.AddPlayer(user, "spectator")
		raid.RegisterChatForStatusUpdates(msgInfo.ChatID, msgInfo.MsgID)
	}()

}

//UpdateUserInfo function
func (raid Raid) UpdateUserInfo(chatID int64, msgID int) {
	keyboard := raid.GetKeyboard()
	raidText := raid.GetText()

	msg := tgbotapi.NewEditMessageText(chatID, msgID, raidText)
	msg.ReplyMarkup = &keyboard
	msg.ParseMode = tgbotapi.ModeHTML
	sender.SendMessage(chatID, msg)
}

//AddPlayer function
func (raid Raid) AddPlayer(user User, role string) {
	str := `INSERT INTO votes (raid_id, user_id, raid_role) VALUES ($1, $2, $3)`
	db.Exec(str, raid, user, role)
}

//GetPlayerRole function
func (raid Raid) GetPlayerRole(user User) (role string, err error) {
	var roles []string
	err = db.Select(&roles, "SELECT raid_role FROM votes WHERE user_id = $1 AND raid_id = $2", user, raid)
	if len(roles) < 1 {
		err = errors.New("not found")
	} else {
		role = roles[0]
	}
	return
}
