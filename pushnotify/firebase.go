package pushnotify

import (
	"context"
	"encoding/json"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	client *messaging.Client
}

type Notification struct {
	Title    string
	Body     string
	ImageURL string

	AppRoute  string
	AppParams map[string]interface{}
}

func New(ConfigFilename string) (*Client, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(ConfigFilename))
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

func (c *Client) SendUserNotification(userID string, notification Notification) error {
	paramsJSONBytes, err := json.Marshal(notification.AppParams)
	if err != nil {
		return err
	}
	paramsJSON := string(paramsJSONBytes)

	_, err = c.client.Send(context.Background(), &messaging.Message{
		Topic: fmt.Sprintf("userid-%s", userID),
		Notification: &messaging.Notification{
			Title:    notification.Title,
			Body:     notification.Body,
			ImageURL: notification.ImageURL,
		},
		Data: map[string]string{
			"appRoute":     notification.AppRoute,
			"appParams":    paramsJSON,
			"topic":        fmt.Sprintf("userid-%s", userID),
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
		},
	})
	return err
}
