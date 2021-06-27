package sms

type Messenger struct {
	apiKey       string
	apiSecret    string
	senderNumber string
}

// Config is used to provide the config required
// to create a Messenger instance.
// APIKey, Secret and SenderNumber are provided
// by the Nexmo developer dashboard, the API Key
// should contain a `api-` prefix, and the sender
// number should be 10 digits (no country-code or
// other symbols should be included).
type Config struct {
	APIKey, Secret, SenderNumber string
}

func New(conf Config) Messenger {
	return Messenger{
		senderNumber: conf.SenderNumber,
		apiKey:       conf.APIKey,
		apiSecret:    conf.Secret,
	}
}

func (m Messenger) SenderNumber() string {
	return m.senderNumber
}
