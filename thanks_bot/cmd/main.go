// Thanks Bot
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	signingSecret string
	db            *gorm.DB
)

func main() {
	signingSecret = os.Getenv("SLACK_SECRET")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", os.Getenv("POSTGRES"), os.Getenv("POSTGRES_PORT"), "postgres", "thanks_bot", os.Getenv("POSTGRES_PASSWORD"), "disable")
	var err error
	db, err = gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Could not connect to postgres: %s", err)
	}
	defer db.Close()
	db.AutoMigrate(&Thank{})

	http.HandleFunc("/thanks", handleNewThanks)
	http.HandleFunc("/mythanks", handleMyThanks)
	http.HandleFunc("/events", handleEvents)
	fmt.Println("[INFO] Server listening")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Couldn't start server: %s\n", err)
	}
}
