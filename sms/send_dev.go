// +build dev

package sms

import (
	"bytes"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ttacon/libphonenumber"
)

var (
	host     = "localhost:5050"
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
	SentAt, Title, From, To, Message, ID string
}

const initalPageTemplate = `<!DOCTYPE html>
<html lang="en">
    <head>
        <title>SMS DevServer</title>
		<link rel="preconnect" href="https://fonts.googleapis.com">
		<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
		<link href="https://fonts.googleapis.com/css2?family=Roboto+Mono&display=swap" rel="stylesheet">
    </head>
    <body>
		<style type="text/css">
		* { margin: 0; padding: 10px; font-family: 'Roboto Mono', monospace; }
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

		</style>

		<div id="fileData">
			<div id="left">
			<h2>Sent SMS (Total {{ len .Messages }})</h2>
			<br/>
			{{ $root := . }}
			{{range $msg := $root.Messages}}
			<div class="msg-item" msg-id="{{$msg.ID}}" onclick="selectMessage('{{$msg.ID}}')">
				<h4>To: {{$msg.To}}</h4>
				<p>{{$msg.Title}}</p>
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
			var conn = new WebSocket("ws://{{.Host}}/ws");
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
<h2>Sent SMS (Total {{ len .Messages }})</h2>
<br/>
{{ $root := . }}
{{range $msg := $root.Messages}}
<div class="msg-item {{if eq $root.Selected $msg.ID }}selected{{end}}" msg-id="{{$msg.ID}}" onclick="selectMessage('{{$msg.ID}}')">
	<h4>To: {{$msg.To}}</h4>
	<p>{{$msg.Title}}</p>
	<p><i>{{$msg.SentAt}}</i></p>
</div>
{{end}}
</div>
<div id="right" class="msg-body">
{{if not .SelectedMessage}}
	<p><i>No message selected.</i></p>
{{else}}
	{{ $msg := .SelectedMessage }}
	<h4>To: {{$msg.To}}, From: {{$msg.From}}, Subject: "{{$msg.Title}}"</h4>
	<br/>
	<p>"</p>
	<p>{{$msg.Message}}</p>
	<p>"</p>
	<br/>
	<p><i>{{$msg.SentAt}}</i></p>
	{{end}}
</div>
`

func init() {
	homeTmpl := template.Must(template.New("").Parse(initalPageTemplate))
	log.Printf("launching SMS dev server @ http://%s/sent_sms\n", host)

	http.HandleFunc("/sent_sms", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		homeTmpl.Execute(res, struct {
			Host     string
			Messages []msg
		}{
			Host:     host,
			Messages: msgs,
		})
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
		if err := http.ListenAndServe(host, nil); err != nil {
			log.Fatal(err)
		}
	}()
}

func (m Messenger) Send(title, to, message string) error {
	toNum, err := libphonenumber.Parse(to, "US")
	if err != nil {
		return err
	}

	msgs = append([]msg{msg{
		ID:      uuid.New().String(),
		SentAt:  time.Now().Format("Jan 02 2006 15:04:05"),
		Title:   title,
		To:      libphonenumber.Format(toNum, libphonenumber.INTERNATIONAL),
		From:    libphonenumber.Format(m.SenderNumber, libphonenumber.INTERNATIONAL),
		Message: message,
	}}, msgs...)

	go func() {
		messagesUpdated <- struct{}{}
	}()

	return nil
}
