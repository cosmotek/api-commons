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
	pipe = make(chan msg)

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
	Messages []msg
}

func createMsgLog() *golive.LiveComponent {
	return golive.NewLiveComponent("msglog", &msglog{})
}

func (t *msglog) Mounted(_ *golive.LiveComponent) {
	go func() {
		for {
			newMsg := <-pipe
			t.Messages = append(t.Messages, newMsg)

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
				<p>Right Half</p>
				<p>Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Vestibulum tortor quam, feugiat vitae, ultricies eget, tempor sit amet, ante. Donec eu libero sit amet quam egestas semper. Aenean ultricies mi vitae est. Mauris placerat eleifend leo. Quisque sit amet est et sapien ullamcorper pharetra. Vestibulum erat wisi, condimentum sed, commodo vitae, ornare sit amet, wisi. Aenean fermentum, elit eget tincidunt condimentum, eros ipsum rutrum orci, sagittis tempus lacus enim ac dui. Donec non enim in turpis pulvinar facilisis. Ut felis. Praesent dapibus, neque id cursus faucibus, tortor neque egestas augue, eu vulputate magna eros eu erat. Aliquam erat volutpat. Nam dui mi, tincidunt quis, accumsan porttitor, facilisis luctus, metus</p>
				<p>Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Vestibulum tortor quam, feugiat vitae, ultricies eget, tempor sit amet, ante. Donec eu libero sit amet quam egestas semper. Aenean ultricies mi vitae est. Mauris placerat eleifend leo. Quisque sit amet est et sapien ullamcorper pharetra. Vestibulum erat wisi, condimentum sed, commodo vitae, ornare sit amet, wisi. Aenean fermentum, elit eget tincidunt condimentum, eros ipsum rutrum orci, sagittis tempus lacus enim ac dui. Donec non enim in turpis pulvinar facilisis. Ut felis. Praesent dapibus, neque id cursus faucibus, tortor neque egestas augue, eu vulputate magna eros eu erat. Aliquam erat volutpat. Nam dui mi, tincidunt quis, accumsan porttitor, facilisis luctus, metus</p>
			</div>
		</div>
	`
}

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
