// Thanks Bot
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nlopes/slack"
)

var (
	signingSecret string
	db            *gorm.DB
)

func informUserOfThanks(toUser, fromUserName, reason string) error {
	api := slack.New(os.Getenv("SLACK_THANKS_BOT_TOKEN"))
	_, _, channelID, err := api.OpenIMChannel(toUser)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("<@%s> just thanked you: %s!", fromUserName, reason)
	api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return nil
}

func handleMyThanks(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if s.Command != "/mythanks" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var timePeriod int64
	timePeriod = 604800
	if s.Text == "hour" || s.Text == "lasthour" || s.Text == "last hour" {
		timePeriod = 3600
	} else if s.Text == "day" || s.Text == "lastday" || s.Text == "last day" || s.Text == "today" {
		timePeriod = 86400
	} else if s.Text == "week"  || s.Text == "lastweek" || s.Text == "last week" {
		timePeriod = 604800
	} else if s.Text == "month" || s.Text == "lastmonth" || s.Text == "last month" {
		timePeriod = 2628000
	} else if s.Text == "quarter" || s.Text == "lastquarter" || s.Text == "last quarter"{
		timePeriod = 7884000
	} else if s.Text == "alltime"  || s.Text == "forever" {
		timePeriod = 9999999999
	}
	fmt.Printf("Retrieving thanks for: %s\n", s.UserID)
	thanks := []Thank{}
	var response string
	if err = db.Where("sent_to = ? AND sent_time > ?", s.UserID, time.Now().UnixNano()-timePeriod*1000000000).Find(&thanks).Error; err != nil {
		fmt.Printf("MyThanks Error: %s\n", err)
		response = "Couldn't fetch your thanks :cry: This is probably not your fault so try again soon!"
	} else {
		if len(thanks) == 0 {
			response = "You haven't received any thanks yet! But surely will soon!"
		} else {
			response = "Here they are! Thanks for being awesome!\n"
		}
		for _, thank := range thanks {
			response = response + "From: *<@" + thank.SentFrom + ">*\n> " + thank.GivenReason + "\n\n"
		}
	}
	params := &slack.Msg{Text: response}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func handleNewThanks(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if s.Command != "/thanks" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return some help text
	if len(s.Text) == 0 {
		params := &slack.Msg{Text: helpText}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}

	mentions, _ := parseMentions(s.Text)
	if len(mentions) < 1 {
		params := &slack.Msg{Text: "You need to give someone to thank!"}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		return
	}
	successfulThanks := make([]string, 0)
	for _, mention := range mentions {
		thanks := Thank{SentFrom: s.UserID, SentTo: mention.ID, SentTime: time.Now().UnixNano(), GivenReason: s.Text, ThankId: randomString(64)}
		thanks.Print()
		if err = db.Create(&thanks).Error; err != nil {
			fmt.Printf("Error creating thanks: %s\n", err)
			continue
		}
		if err = informUserOfThanks(mention.ID, s.UserID, s.Text); err != nil {
			fmt.Printf("Couldn't send message to user: %s\n", err)
			continue
		}
		successfulThanks = append(successfulThanks, mention.ID)
	}
	mentionString := "<@" + strings.Join(successfulThanks, ">, <@") + ">"
	params := &slack.Msg{Text: "Thanks have been sent to: " + mentionString + "!"}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

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
	fmt.Println("[INFO] Server listening")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Couldn't start server: %s\n", err)
	}
}
