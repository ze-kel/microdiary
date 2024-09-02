package db

import (
	"os"

	"gorm.io/driver/postgres"
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

type FilesForMessages struct {
	gorm.Model
	MessageId int64  `gorm:"primaryKey;not null"`
	ChatId    int64  `gorm:"not null"`
	Filename  string `gorm:"not null"`
	Format    string `gorm:"not null"`
	Date      int64  `gorm:"not null"`
}

type Tabler interface {
	TableName() string
}

func (Message) TableName() string {
	return "Microdiary_Messages"
}

func (FilesForMessages) TableName() string {
	return "Microdiary_Files"
}

func Init(pgUrl string) *gorm.DB {

	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		panic("failed to make ./db directory")
	}

	var dbConnection gorm.Dialector

	if len(pgUrl) > 0 {
		dbConnection = postgres.Open(pgUrl)
	} else {
		dbConnection = sqlite.Open("./db/messages.db")
	}

	db, err := gorm.Open(dbConnection, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Message{})
	db.AutoMigrate(&FilesForMessages{})

	return db
}
