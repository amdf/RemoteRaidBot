package main

var chRaids = make(chan Raid, 100)

func initialUpdateRaids() {
	var raids []Raid

	errRaids := db.Select(&raids, "SELECT DISTINCT raid_id FROM raids")
	if errRaids == nil {
		for _, raid := range raids {
			chRaids <- raid
		}
	}
}

func updateUserInfo() {
	for raid := range chRaids {
		raidText := raid.GetText()

		//first, update admin info
		raid.UpdateRaidAdminInfo()

		//second, update all other chats
		var messages []ChatAndMessage
		err := db.Select(&messages, "SELECT chat_id, msg_id FROM chats WHERE raid_id = $1", raid)

		if err == nil {
			for _, msg := range messages {
				raid.UpdateUserInfo(msg.ChatID, msg.MsgID, raidText)
			}
		}

		//third, update public chat messages
		var groupMessages []string
		err = db.Select(&groupMessages, "SELECT inline_msg_id FROM groupmessages WHERE raid_id = $1", raid)

		if err == nil {
			for _, inlineMsgID := range groupMessages {
				raid.UpdatePublicInfo(inlineMsgID, raidText)
			}
		}
		finished, _ := raid.IsFinished()
		if finished {
			raid.Delete()

		}
	}
}

//GetAdminChats returns admin's chat IDs (for active raids)
func GetAdminChats() (adminChats []int64, err error) {
	var adminUsers []User
	err = db.Select(&adminUsers, "SELECT DISTINCT user_id FROM raids")
	for _, x := range adminUsers {
		adminChats = append(adminChats, int64(x))
	}
	return
}

//GetVotersChats returns voter's chat IDs (for active raids)
func GetVotersChats() (votersChats []int64, err error) {
	var voters []User
	err = db.Select(&voters, "SELECT DISTINCT user_id FROM votes")
	for _, x := range voters {
		votersChats = append(votersChats, int64(x))
	}
	return
}

//GetAllChats returns all chat IDs (for active raids)
func GetAllChats() (allChats []int64, err error) {
	err = db.Select(&allChats, "SELECT DISTINCT chat_id FROM chats")
	return
}

//GetSubscribers returns users who wants to receive invites to new raids
func GetSubscribers() (subscribers []User, err error) {
	err = db.Select(&subscribers, "SELECT DISTINCT user_id FROM players WHERE notif = true")
	return
}
