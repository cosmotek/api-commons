package email

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Sender string

type Messenger struct {
	apiKey       string
	senderDomain string
}

type Config struct {
	// APIKey should look like `api-<some hash>`
	APIKey, SenderDomain string
}

func New(conf Config) Messenger {
	return Messenger{
		senderDomain: conf.SenderDomain,
		apiKey:       conf.APIKey,
	}
}

func (m Messenger) GenerateSender(name, emailUser string) Sender {
	return Sender(fmt.Sprintf("%s <%s@%s>", name, emailUser, m.senderDomain))
}

func handleResponse(res *http.Response) error {
	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to complete error build, could not read response body: %s", err.Error())
		}

		return fmt.Errorf("failed to send email: status_code:%d, res_body:'%s'", res.StatusCode, string(resBody))
	}

	return nil
}

func (m Messenger) url() string {
	return fmt.Sprintf(
		"https://api:%s@api.mailgun.net/v3/%s/messages",
		m.apiKey,
		m.senderDomain,
	)
}
