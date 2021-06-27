package main

import (
	"time"

	"github.com/cosmotek/api-commons/sms"
)

func main() {
	messenger := sms.New(sms.Config{})

	for {
		time.Sleep(time.Second * 3)

		messenger.Send(time.Now().String(), "7408772320", "Im great")
	}
}
