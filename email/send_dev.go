// +build dev

package email

import (
	"bytes"
	"html"
	"io"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"goji.io"
	"goji.io/pat"
)

var (
	host     = "localhost:5051"
	newline  = []byte{'\n'}
	space    = []byte{' '}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	messagesUpdated = make(chan struct{})
	msgs            = []msg{}
)

type msg struct {
	SentAt, Subject, From, To, Text, HTML, ID string
}

const initalPageTemplate = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>Email DevServer</title>
		<link rel="preconnect" href="https://fonts.googleapis.com">
		<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
		<link href="https://fonts.googleapis.com/css2?family=Roboto+Mono&display=swap" rel="stylesheet">
    </head>
    <body>
		<style type="text/css">
		* { margin: 0; padding: 10px; font-family: 'Roboto Mono', monospace; background-color: white; }
		#left { position: absolute; left: 0; top: 0; max-width: 30%; height: 100vh; overflow-y: scroll; }
		#right { position: absolute; right: 0; top: 0; width: 65%; height: 100vh; overflow-y: scroll; }

		.msg-item {
			margin-bottom: 10px;
			border-bottom: solid 1px grey;
			padding: 5px;
			font-family: 'Noto Sans', sans-serif;
			overflow: hidden;
		}

		.selected {
			background-color: #ADD8E6;
			border-radius: 10px;
		}

		iframe {
			display: block;
			width: 100%;
			height: 100%;
			border: none;
		}
		</style>

		<div id="fileData">
			<div id="left">
			<h2>Sent Emails (Total {{ len .Messages }})</h2>
			<br/>
			{{ $root := . }}
			{{range $msg := $root.Messages}}
			<div class="msg-item" msg-id="{{$msg.ID}}" onclick="selectMessage('{{$msg.ID}}')">
				<h4>To: {{$msg.To}}</h4>
				<p>{{$msg.Subject}}</p>
				<p><i>{{$msg.SentAt}}</i></p>
			</div>
			{{end}}
			</div>
			<div id="right" class="msg-body">
				<p><i>No message selected.</i></p>
			</div>
		</div>

        <script type="text/javascript">
			var data = document.getElementById("fileData");
			var conn = new WebSocket("ws://{{.Host}}/email_ws");
			conn.onmessage = function(evt) {
				data.innerHTML = evt.data;
			}

			function selectMessage(id) {
				conn.send(id);
			}
        </script>
    </body>
</html>
`

const updatedPageTemplate = `
<div id="left">
<h2>Sent Emails (Total {{ len .Messages }})</h2>
<br/>
{{ $root := . }}
{{range $msg := $root.Messages}}
<div class="msg-item {{if eq $root.Selected $msg.ID }}selected{{end}}" msg-id="{{$msg.ID}}" onclick="selectMessage('{{$msg.ID}}')">
	<h4>To: {{$msg.To}}</h4>
	<p>{{$msg.Subject}}</p>
	<p><i>{{$msg.SentAt}}</i></p>
</div>
{{end}}
</div>
<div id="right" class="msg-body">
{{if not .SelectedMessage}}
	<p><i>No message selected.</i></p>
{{else}}
	{{ $msg := .SelectedMessage }}
	<h4><xmp>To: {{$msg.To}}, From: {{$msg.From}}, Subject: "{{$msg.Subject}}"</xmp></h4>
	<br/>
	<p><b>HTML</b></p>
	<iframe src="/sent_emails/body/{{$msg.ID}}"/>
	<p><b>Plain Text</b></p>
	<p>"</p>
	<xmp>{{$msg.Text}}</xmp>
	<p>"</p>
	<br/>
	<p><i>{{$msg.SentAt}}</i></p>
	{{end}}
</div>
`

func init() {
	homeTmpl := template.Must(template.New("").Parse(initalPageTemplate))
	log.Printf("launching SMS dev server @ http://%s/sent_emails\n", host)
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/sent_emails"), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		homeTmpl.Execute(res, struct {
			Host     string
			Messages []msg
		}{
			Host:     host,
			Messages: msgs,
		})
	})

	mux.HandleFunc(pat.Get("/sent_emails/body/:id"), func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		id := pat.Param(req, "id")

		if len(id) > 0 {
			for _, m := range msgs {
				if m.ID == id {
					io.WriteString(res, m.HTML)
					break
				}
			}
		}
	})

	mux.HandleFunc(pat.New("/email_ws"), func(w http.ResponseWriter, r *http.Request) {
		selectedMsgID := ""
		selectedMsg := new(msg)
		selectedMsg = nil
		selectionChanged := make(chan string)

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Println(err)
			}
			return
		}

		go func() {
			for {
				select {
				case <-messagesUpdated:
				case selection := <-selectionChanged:
					selectedMsgID = selection
					if len(selection) > 0 {
						for _, m := range msgs {
							if m.ID == selection {
								selectedMsg = &m
								break
							}
						}
					}
				}

				buff := bytes.NewBuffer([]byte{})

				template.Must(template.New("").Parse(updatedPageTemplate)).Execute(buff, struct {
					Messages        []msg
					Selected        string
					SelectedMessage *msg
				}{
					Messages:        msgs,
					Selected:        selectedMsgID,
					SelectedMessage: selectedMsg,
				})

				ws.SetWriteDeadline(time.Now().Add(time.Second * 5))
				if err := ws.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
					return
				}
			}
		}()

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}

				ws.Close()
				break
			}

			go func() {
				selectionChanged <- string(bytes.TrimSpace(bytes.Replace(message, newline, space, -1)))
			}()
		}
	})

	go func() {
		if err := http.ListenAndServe(host, mux); err != nil {
			log.Fatal(err)
		}
	}()
}

func (m Messenger) Send(subject, to, text string, from Sender) error {
	msgs = append([]msg{msg{
		ID:      uuid.New().String(),
		SentAt:  time.Now().Format("Jan 02 2006 15:04:05"),
		Subject: subject,
		To:      to,
		From:    string(from),
		Text:    text,
	}}, msgs...)

	go func() {
		messagesUpdated <- struct{}{}
	}()

	return nil
}

func (m Messenger) SendHTML(subject, to, htmlBody string, from Sender) error {
	msgs = append([]msg{msg{
		ID:      uuid.New().String(),
		SentAt:  time.Now().Format("Jan 02 2006 15:04:05"),
		Subject: subject,
		To:      to,
		From:    string(from),
		HTML:    htmlBody,
		Text:    html.EscapeString(htmlBody),
	}}, msgs...)

	go func() {
		messagesUpdated <- struct{}{}
	}()

	return nil
}
