package vault

import (
	"github.com/hashicorp/vault/api"
)

// Vault is a simple wrapper around the Go Vault client
// which handles authentication (AppRole) auto-magically.
type Vault struct {
	client *api.Client
}

// Config is used by the Dial function to connect with Vault.
type Config struct {
	Address string
	Token   string
}

// Dial connects to and authenticates with Vault using AppRole
// provided by config.
func Dial(config Config) (*Vault, error) {
	conf := api.DefaultConfig()
	conf.Address = config.Address

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	client.SetToken(config.Token)
	return &Vault{client}, nil
}

// Logical returns the logical client provided by the raw
// Vault Go library (which can be used to make read/write/delete
// calls). See https://github.com/hashicorp/vault/api for details.
func (v *Vault) Logical() *api.Logical {
	return v.client.Logical()
}
