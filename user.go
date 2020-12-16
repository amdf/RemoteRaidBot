package main

import "fmt"

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
