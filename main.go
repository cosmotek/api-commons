package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cosmotek/api-commons/sms"
)

func main() {
	messenger := sms.New(sms.Config{})
	count := 0

	for {
		time.Sleep(time.Second * 3)

		messenger.Send(time.Now().String(), "7408772320", fmt.Sprintf("count is %d", count))
		log.Printf("count is %d\n", count)
		count++
	}
}
