package db

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	MessageId int64  `gorm:"primaryKey;not null"`
	ChatId    int64  `gorm:"not null"`
	Message   string `gorm:"not null"`
	Date      int64  `gorm:"not null"`
}

func Init() *gorm.DB {

	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		panic("failed to make ./db directory")
	}

	db, err := gorm.Open(sqlite.Open("./db/messages.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Message{})

	return db
}
