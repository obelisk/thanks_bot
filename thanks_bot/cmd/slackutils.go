// Thanks Bot
package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	return randomStringWithCharset(length, charset)
}

func parseEscapedMention(mentionStr string) (SlackMention, error) {
	mention := SlackMention{}
	if len(mentionStr) == 0 {
		return mention, fmt.Errorf("Mention must have non zero length")
	}

	parts := strings.Split(mentionStr, "|")
	// If we have the form <@UXXXXXXXX|name>, take the first element split on |
	// <@UXXXXXXXX and append > to convert it to the new format. Otherwise the
	// split will do nothing and it should already be in that form
	converted := parts[0]
	if len(parts) > 1 {
		converted = parts[0] + ">"
	}

	if converted[1] != '@' || converted[1] == '#' {
		return mention, fmt.Errorf("Cannot parse, mention type unknown")
	}

	if converted[1] == '@' {
		mention.Type = "user"
	} else {
		mention.Type = "channel"
	}

	mention.ID = converted[2:len(converted)-1]
	mention.Name = ""
	return mention, nil
}

func parseMentions(msg string) (mentions []SlackMention, errors []error) {
	// This regex is more complicated than it needs to be to still support
	// the old <@XXXXXXX|name> syntax. This is deprecated by is still used
	// by the commands API. The events API uses the more modern <@UXXXX>
	userRegex := "<@[A-Z0-9]+(\\|[a-zA-Z0-9][a-zA-Z0-9.\\-_]*)*>"
	foundIds := make(map[string]bool, 0)
	r := regexp.MustCompile(userRegex)
	for _, mentionIndexes := range r.FindAllStringSubmatchIndex(msg, -1) {
		m, err := parseEscapedMention(msg[mentionIndexes[0]:mentionIndexes[1]])
		if err != nil {
			errors = append(errors, err)
		} else {
			if _, ok := foundIds[m.ID]; !ok {
				mentions = append(mentions, m)
				foundIds[m.ID] = true
			}
		}
	}
	return
}

