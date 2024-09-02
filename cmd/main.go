package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	mdbot "github.com/ze-kel/microdiary/cmd/bot"
	"github.com/ze-kel/microdiary/cmd/db"
	storage "github.com/ze-kel/microdiary/cmd/filestorage"
)

func init() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Print("WARN: No .env file found. This is fine if you are running inside docker container.")
	}
}

func initBot() {
	token, exists := os.LookupEnv("TG_TOKEN")
	if !exists {
		panic("NO TOKEN IN ENV")
	}

	pgUrl, isPg := os.LookupEnv("POSTGRES_URL")

	if isPg {
		log.Printf("POSTGRES_URL is set, connecting to external db")
	} else {
		log.Print("POSTGRES_URL is not set, using local sqlite")
	}

	database := db.Init(pgUrl)

	var s storage.GenericStorage

	bucketUrl, isBurl := os.LookupEnv("BUCKET_URL")
	bucketName, isBname := os.LookupEnv("BUCKET_NAME")
	bucketCredId, isBCI := os.LookupEnv("BUCKET_CRED_ID")
	bucketCredSecret, isBCS := os.LookupEnv("BUCKET_CRED_SECRET")

	if isBurl && isBname && isBCI && isBCS {
		log.Printf("BUCKET_URL is set, using s3 storage")
		s = storage.NewS3(context.Background(), storage.S3Params{
			BucketName: bucketName,
			Url:        bucketUrl,
			ID:         bucketCredId,
			Secret:     bucketCredSecret,
		})
	} else {

		log.Printf("BUCKET_URL, BUCKET_NAME, BUCKET_CRED_ID or BUCKET_CRED_SECRET is not set, storing files in /db")
		s = storage.NewLocal("./db")
	}

	m := mdbot.New(database, token, s)

	m.Start()
}

func initHealthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Print("Running healthcheck on :80 /health")

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println("Error starting http server to provide healthcheck. Bot might be running", err)
	}

}

func main() {
	go initBot()
	initHealthCheck()
}
