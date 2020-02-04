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

func parseEscapedMention(mentionStr string) (mention SlackMention, err error) {
	parts := strings.Split(mentionStr, "|")
	if len(parts) != 2 {
		err = fmt.Errorf("Cannot parse, error with '|'")
		return
	}
	if len(parts[0]) < 2 {
		err = fmt.Errorf("Cannot parse, user id invalid")
		return
	}
	if parts[0][1] != '@' || parts[0][1] == '#' {
		err = fmt.Errorf("Cannot parse, mention type unknown")
		return
	}
	if parts[0][1] == '@' {
		mention.Type = "user"
	} else {
		mention.Type = "channel"
	}
	mention.ID = parts[0][2:]
	if len(parts[1]) == 0 {
		err = fmt.Errorf("Cannot parse, name is length zero")
		return
	}
	mention.Name = strings.TrimSuffix(parts[1], ">")
	return
}

func parseMentions(msg string) (mentions []SlackMention, errors []error) {
	userRegex := "<@[A-Z0-9]*\\|[a-zA-Z.\\-_]*>"
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

