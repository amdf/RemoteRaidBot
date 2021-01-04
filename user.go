package main

import (
	"database/sql"
	"fmt"
	"log"
)

//User - telegram and bot user
type User int

//IsRegistered bool
func (userID User) IsRegistered() (result bool) {
	rows, err := db.Query("SELECT user_id FROM players where user_id = $1", userID)

	if err == nil {
		result = rows.Next()
		rows.Close()
	} else {
		fmt.Println(err)
	}
	return
}

//IsNotificationsEnabled bool
func (userID User) IsNotificationsEnabled() (result bool) {
	rows, err := db.Query("SELECT notif FROM players where user_id = $1", userID)

	if err == nil {
		if rows.Next() {
			rows.Scan(&result)
		}
		rows.Close()
	} else {
		fmt.Println(err)
	}
	return
}

//SetName function
func (userID User) SetName(name string) {
	rs := []rune(name)
	if len(rs) > 255 {
		name = string(rs[:255])
	}
	if !userID.IsRegistered() {
		str := `INSERT INTO players (user_id, pogoname, notif) VALUES ($1, $2, $3)`
		db.Exec(str, userID, name, false)
	} else {
		str := `UPDATE players SET pogoname = $1 where user_id = $2`
		db.Exec(str, name, userID)
		userID.updateRaids()
	}
}

//SetCode function to store code
func (userID User) SetCode(pogoCode string) bool {
	rs := []rune(pogoCode)

	var digitsOnly []rune
	for _, r := range rs {
		if r == '0' || r == '1' || r == '2' || r == '3' || r == '4' || r == '5' || r == '6' || r == '7' || r == '8' || r == '9' {
			digitsOnly = append(digitsOnly, r)
		}
	}

	if len(digitsOnly) > 12 {
		digitsOnly = digitsOnly[:12]
	}

	pogoCode = string(digitsOnly)
	str := `UPDATE players SET pogocode = $1 where user_id = $2`
	db.Exec(str, pogoCode, userID)

	userID.updateRaids()

	return (len(digitsOnly) == 12)
}

//GetRaids function
func (userID User) GetRaids() (raids []Raid, err error) {
	err = db.Select(&raids, "SELECT DISTINCT raid_id FROM votes WHERE user_id = $1", userID)
	return
}

//GetCode - pokemon go friend code
func (userID User) GetCode() (result string) {
	var pogoCode sql.NullString
	rows, err := db.Queryx("SELECT pogocode FROM players where user_id = $1", userID)

	if err == nil {
		if rows.Next() {
			err = rows.Scan(&pogoCode)
		}
		rows.Close()
	}

	if pogoCode.Valid {
		result = pogoCode.String
	} else {
		result = "не задан"
	}

	return
}

//GetName - registered user name
func (userID User) GetName() (result string) {

	rows, err := db.Query("SELECT pogoname FROM players WHERE user_id = $1", userID)

	if err == nil {
		if rows.Next() {
			err = rows.Scan(&result)
		}
		rows.Close()
	}

	if "" == result {
		result = fmt.Sprintf("user%d", userID)
	}

	return
}

//Unregister function
func (userID User) Unregister() {
	str := `DELETE FROM players WHERE user_id = $1`

	db.Exec(str, userID)
}

//EnableNotifications function
func (userID User) EnableNotifications(enable bool) {
	str := `UPDATE players SET notif = $1 WHERE user_id = $2`
	db.Exec(str, enable, userID)
}

func (userID User) updateRaids() {
	raids, err := userID.GetRaids()
	if err == nil {
		for _, r := range raids {
			infoUpdated[r] = false
		}
	}
}

//DeleteCode function
func (userID User) DeleteCode() {
	str := `UPDATE players SET pogocode = NULL WHERE user_id = $1`
	db.Exec(str, userID)
	userID.updateRaids()
}

//Vote function
func (userID User) Vote(raid Raid, role string) {
	var changed bool
	_, err := raid.GetPlayerRole(userID)
	if err == nil {
		str := `UPDATE votes SET raid_role = $1 WHERE user_id = $2 AND raid_id = $3 AND raid_role <> $4 RETURNING user_id`

		rows, err := db.Query(str, role, userID, raid, role)
		if err == nil {
			changed = rows.Next()
			rows.Close()
		}

	} else {
		if userID.IsRegistered() {
			if raid.Started() {
				raid.AddPlayer(userID, role)
				changed = true
			}
		}
	}

	if changed {
		log.Println("changed")
		infoUpdated[raid] = false
	} else {
		log.Println("not changed")
	}
}
