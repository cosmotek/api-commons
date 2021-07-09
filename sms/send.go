// +build !dev

package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type nexmoSendResponse struct {
	Messages []struct {
		Status string `json:"status"`
	} `json:"messages"`
}

func (m Messenger) Send(title, to, message string) error {
	buff := bytes.Buffer{}
	err := json.NewEncoder(&buff).Encode(map[string]interface{}{
		"api_key":    m.apiKey,
		"api_secret": m.apiSecret,
		"to":         fmt.Sprintf("1%s", to),
		"from":       fmt.Sprintf("1%s", m.SenderNumber.String()),
		"text":       message,
		"title":      title,
	})
	if err != nil {
		return err
	}

	res, err := http.Post("https://rest.nexmo.com/sms/json", "application/json", &buff)
	if err != nil {
		return err
	}

	nexmoRes := nexmoSendResponse{}
	err = json.NewDecoder(res.Body).Decode(&nexmoRes)
	if err != nil {
		return err
	}

	if nexmoRes.Messages[0].Status == "0" {
		return nil
	}

	statusCode, err := strconv.ParseInt(nexmoRes.Messages[0].Status, 10, 64)
	if err != nil {
		return err
	}

	msg := errorCodeToMessage(int(statusCode))
	if msg == "" {
		return fmt.Errorf(
			"unknown server error: failed to send sms, recipient=%s, status_code=%s",
			fmt.Sprintf("1%s", to),
			nexmoRes.Messages[0].Status,
		)
	}

	return fmt.Errorf(
		"coded server error: failed to send sms, recipient=%s, status_code=%d, error_msg=%s",
		fmt.Sprintf("1%s", to),
		statusCode,
		msg,
	)
}
