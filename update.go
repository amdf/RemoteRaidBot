package main

import (
	"database/sql"
	"fmt"
	"time"
)

//AdminRaidInfo struct
type AdminRaidInfo struct {
	RaidInfo  string        `db:"raid_info"`
	AdminName string        `db:"admin_name"`
	ChatID    int64         `db:"chat_id"`
	MsgID     sql.NullInt64 `db:"msg_id"`
}

//AdminRaidHeader contains info about raid
type AdminRaidHeader struct {
	RaidID int `db:"raid_id"`
	AdminRaidInfo
}

var cachedAdminRaidInfo = make(map[int]AdminRaidInfo)

func updateAdminsInfo() {
	for {
		var rh []AdminRaidHeader
		err := db.Select(&rh, "SELECT raid_id, raid_info, pogoname AS admin_name, chat_id, msg_id FROM raids LEFT JOIN players USING (user_id)")

		if err == nil {
			for _, raidHeader := range rh {
				cachedRH, ok := cachedAdminRaidInfo[raidHeader.RaidID]
				if ok {
					if cachedRH != raidHeader.AdminRaidInfo {
						fmt.Printf("raid %d changed!\r\n", raidHeader.RaidID)
						cachedAdminRaidInfo[raidHeader.RaidID] = raidHeader.AdminRaidInfo //update cache

						if raidHeader.MsgID.Valid { //has admin interface
							fmt.Printf("send update to raid %d\r\n", raidHeader.RaidID)
							sender.EditText(raidHeader.ChatID, int(raidHeader.MsgID.Int64), raidHeader.RaidInfo) //send update to admin
						} else {
							fmt.Printf("raid %d is not yet started\r\n", raidHeader.RaidID)
						}
					}
				} else {
					fmt.Printf("make cache for raid %d\r\n", raidHeader.RaidID)
					cachedAdminRaidInfo[raidHeader.RaidID] = raidHeader.AdminRaidInfo
				}
			}
		} else {
			fmt.Printf("error get raid info: %s\r\n", err.Error())
		}
		time.Sleep(time.Second)
	}
}

//GetAdminChats returns admin's chat IDs (for active raids)
func GetAdminChats() (adminChats []int64, err error) {
	err = db.Select(&adminChats, "SELECT DISTINCT chat_id FROM raids")
	return
}

//GetVotersChats returns voter's chat IDs (for active raids)
func GetVotersChats() (votersChats []int64, err error) {
	err = db.Select(&votersChats, "SELECT DISTINCT chat_id FROM votes")
	return
}

//GetAllChats returns all chat IDs (for active raids)
func GetAllChats() (allChats []int64, err error) {
	err = db.Select(&allChats, "SELECT DISTINCT chat_id FROM raids FULL OUTER JOIN votes USING (chat_id)")
	return
}
