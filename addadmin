package main

import(
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Session struct {
	UserID	int `gorm:"primary_key"`
	ChatID	int64
	IsAdmin	bool
	State	uint
	Timeout	uint
}


func main() {
	db, err = gorm.Open("sqlite3", "server.db")
	if err != nil {
		panic("Failed to connect to database")
	}

	defer db.Close()
}
