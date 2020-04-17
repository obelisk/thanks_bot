// Thanks Bot
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/slack-go/slack"
)

func informUserOfThanks(toUser, fromUserName, reason string) error {
	return messageUser(toUser, fmt.Sprintf("<@%s> just thanked you! They said: %s", fromUserName, reason))
}

func messageUser(toUser, message string) error {
	api := slack.New(os.Getenv("SLACK_THANKS_BOT_TOKEN"))
	_, _, channelID, err := api.OpenIMChannel(toUser)
	if err != nil {
		return err
	}
	api.PostMessage(channelID, slack.MsgOptionText(message, false))
	return nil
}

func processThanks(from, message string) (string, error) {
	mentions, _ := parseMentions(message)
	if len(mentions) < 1 {
		return "You need to give someone to thank!", nil
	}
	successfulThanks := make([]string, 0)
	for _, mention := range mentions {
		// Disallow Self Thanks
		if from == mention.ID {
			//continue
		}

		thanks := Thank{
			SentFrom:    from,
			SentTo:      mention.ID,
			SentTime:    time.Now().UnixNano(),
			GivenReason: message,
			ThankId:     randomString(64),
		}

		thanks.Print()
		if err := db.Create(&thanks).Error; err != nil {
			fmt.Printf("Error creating thanks: %s\n", err)
			// Make this a continue incase the error is ephemeral.
			// Someone isn't going to get thanked but the user will
			// be able to try again.
			continue
		}
		if err := informUserOfThanks(mention.ID, from, message); err != nil {
			fmt.Printf("Couldn't send message to user: %s\n", err)
			// This is likely a slack misconfig or you can't send a message
			// to this user I.e ThanksBot messaging itself.
			continue
		}
		successfulThanks = append(successfulThanks, mention.ID)
	}
	if len(successfulThanks) == 0 {
		return "You can't thank those people! Sorry!", nil
	}
	mentionString := "<@" + strings.Join(successfulThanks, ">, <@") + ">"
	return "Thanks have been sent to: " + mentionString + "!", nil
}

