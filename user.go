package main

import (
	"database/sql"
	"fmt"
)

//User - telegram and bot user
type User int

//IsRegistered bool
func (user_id User) IsRegistered() (result bool) {
	rows, err := db.Query("SELECT user_id FROM players where user_id = $1", user_id)

	if err == nil {
		result = rows.Next()
		rows.Close()
	} else {
		fmt.Println(err)
	}
	return
}

//IsNotificationsEnabled bool
func (user_id User) IsNotificationsEnabled() (result bool) {
	rows, err := db.Query("SELECT notif FROM players where user_id = $1", user_id)

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

//Register function
func (user_id User) Register(pogoName string) {
	rs := []rune(pogoName)
	if len(rs) > 255 {
		pogoName = string(rs[:255])
	}
	str := `INSERT INTO players (user_id, pogoname, notif) VALUES ($1, $2, $3)`
	db.Exec(str, user_id, pogoName, true)
}

//SetCode function to store code
func (user_id User) SetCode(pogoCode string) bool {
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
	db.Exec(str, pogoCode, user_id)

	return (len(digitsOnly) == 12)
}

//GetCode - pokemon go friend code
func (user_id User) GetCode() (result string) {
	var pogoCode sql.NullString
	rows, err := db.Queryx("SELECT pogocode FROM players where user_id = $1", user_id)

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

//Unregister function
func (user_id User) Unregister() {
	str := `DELETE FROM players WHERE user_id = $1`

	db.Exec(str, user_id)
}

//EnableNotifications function
func (user_id User) EnableNotifications(enable bool) {
	str := `UPDATE players SET notif = $1 WHERE user_id = $2`
	db.Exec(str, enable, user_id)
}
