// Thanks Bot
package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type Thank struct {
	gorm.Model
	SentTo      string
	SentFrom    string
	GivenReason string
	ThankId     string
	SentTime    int64
}

type SlackMention struct {
	Type string
	ID   string
	Name string
}

func (thanks Thank) Print() {
	fmt.Printf("To: %s\n", thanks.SentTo)
	fmt.Printf("From: %s\n", thanks.SentFrom)
	fmt.Printf("Reason: %s\n", thanks.GivenReason)
	fmt.Printf("Time: %d\n", thanks.SentTime)
	fmt.Printf("Thank ID: %s\n", thanks.ThankId)
}
