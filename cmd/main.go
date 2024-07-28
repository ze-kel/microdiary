package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	mdbot "github.com/ze-kel/microdiary/cmd/bot"
	"github.com/ze-kel/microdiary/cmd/db"
)

func init() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Print("WARN: No .env file found. This is fine if you are running inside docker container.")
	}
}

func main() {
	token, exists := os.LookupEnv("TG_TOKEN")
	if !exists {
		panic("NO TOKEN IN ENV")
	}
	database := db.Init()
	m := mdbot.New(database, token)

	m.Start()
}
