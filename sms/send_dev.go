// +build dev

package sms

import (
	"fmt"
	"log"
	"time"

	"github.com/brendonmatos/golive"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var pipe chan msg

func init() {
	pipe = make(chan msg, 100)

	app := fiber.New()
	liveServer := golive.NewServer()

	app.Get("/sent_sms", liveServer.CreateHTMLHandler(createMsgLog, golive.PageContent{
		Lang:  "us",
		Title: "SMS Dev Server",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	go app.Listen(":8080")

	log.Println("launching SMS dev server @ http://localhost:8080/sent_sms")
}

type msglog struct {
	golive.LiveComponentWrapper
	Messages        []msg
	SelectedMessage *msg
}

func createMsgLog() *golive.LiveComponent {
	return golive.NewLiveComponent("msglog", &msglog{})
}

func (t *msglog) Mounted(_ *golive.LiveComponent) {
	go func() {
		for {
			newMsg := <-pipe
			t.Messages = append([]msg{newMsg}, t.Messages...)

			t.Commit()
		}
	}()
}

func (t *msglog) TemplateHandler(_ *golive.LiveComponent) string {
	return `
		<div>
			<title>Some Example from CSS-Tricks</title>
			<style type="text/css">
			* { margin: 0; padding: 0; }
			p { padding: 10px; }
			#left { position: absolute; left: 0; top: 0; width: 25%; }
			#right { position: absolute; right: 0; top: 0; width: 75%; }

			</style>

			<div id="left">
			<h2>Sent SMS</h2>
			<br/>
			{{range $msg := .Messages}}
			<h3>{{$msg.Title}}</h3>
			<p>{{$msg.Message}}</p>
			<p><i>{{$msg.SentAt}}</i></p>
			{{end}}
			</div>

			<div id="right">
			{{if not .SelectedMessage}}
				<p><i>No message selected.</i></p>
			{{else}} 
				<p>testy</p>
			{{end}}
			</div>
		</div>
	`
}

// <h3>{{.SelectedMessage.Title}}</h3>
// 				<p>{{.SelectedMessage.Message}}</p>
// 				<p><i>{{.SelectedMessage.SentAt}}</i></p>

type msg struct {
	SentAt                   time.Time
	Title, From, To, Message string
}

func (m Messenger) Send(title, to, message string) error {
	pipe <- msg{
		SentAt:  time.Now(),
		Title:   title,
		To:      fmt.Sprintf("1%s", to),
		From:    fmt.Sprintf("1%s", m.senderNumber),
		Message: message,
	}

	return nil
}
