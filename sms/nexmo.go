package sms

import "github.com/ttacon/libphonenumber"

type Messenger struct {
	apiKey       string
	apiSecret    string
	SenderNumber *libphonenumber.PhoneNumber
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

func New(conf Config) (Messenger, error) {
	fromNum, err := libphonenumber.Parse(conf.SenderNumber, "US")
	if err != nil {
		return Messenger{}, err
	}

	return Messenger{
		SenderNumber: fromNum,
		apiKey:       conf.APIKey,
		apiSecret:    conf.Secret,
	}, nil
}
