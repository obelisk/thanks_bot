// Thanks Bot
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

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

	returnMessage, err := processThanks(s.UserID, s.Text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := &slack.Msg{Text: returnMessage}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: os.Getenv("SLACK_VERIFICATION_TOKEN")}))
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			fmt.Printf("%s\n", ev.Text)
			message, err := processThanks(ev.User, ev.Text)
			if err != nil {
				fmt.Printf("Internal Error: %s", err)
				message = "ThanksBot is experiencing some issues! Please try again later!"
			}
			if err = messageUser(ev.User, message); err != nil {
				fmt.Printf("Internal Error %s", err)
			}
		}
	}
}
