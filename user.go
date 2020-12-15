package main

import "fmt"

//User - telegram and bot user
type User int

//IsRegistered bool
func (user_id User) IsRegistered() (result bool) {
	rows, err := db.Query("SELECT userid FROM players where userid = $1", user_id)

	if err == nil {
		result = rows.Next()
		rows.Close()
	} else {
		fmt.Println(err)
	}
	return
}

//Register function
func (user_id User) Register(pogoName string) {
	if len(pogoName) > 255 {
		rs := []rune(pogoName)
		pogoName = string(rs[:255])
	}
	str := `INSERT INTO players (userid, pogoname, notif) VALUES ($1, $2, $3)`
	db.Exec(str, user_id, pogoName, true)
}

//Unregister function
func (user_id User) Unregister() {
	str := `DELETE FROM players WHERE userid = $1`

	db.Exec(str, user_id)
}
