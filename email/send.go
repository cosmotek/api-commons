// +build !dev

package email

import (
	"net/http"
	"net/url"
)

func (m Messenger) Send(subject, to, text string, from Sender) error {
	res, err := http.PostForm(m.url(), url.Values{
		"from":    {string(from)},
		"to":      {to},
		"subject": {subject},
		"text":    {text},
	})

	if err != nil {
		return err
	}

	return handleResponse(res)
}

func (m Messenger) SendHTML(subject, to, html string, from Sender) error {
	res, err := http.PostForm(m.url(), url.Values{
		"from":    {string(from)},
		"to":      {to},
		"subject": {subject},
		"html":    {html},
	})
	if err != nil {
		return err
	}

	return handleResponse(res)
}
