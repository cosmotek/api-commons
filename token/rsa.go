package token

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
)

func ParseRSAKeys(pubKeyFile, privateKeyFile string) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	privateKeyBytes, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
	if err != nil {
		return nil, nil, err
	}

	pubkey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	return pubkey, privateKey, nil
}

func ParsePublicRSAKey(pubKeyFile string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := ioutil.ReadFile(pubKeyFile)
	if err != nil {
		return nil, err
	}

	pubkey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}
