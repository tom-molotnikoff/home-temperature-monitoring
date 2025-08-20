package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OAUTH_TOKEN *oauth2.Token
var OAUTH_SET = false

func getTokenSource() (oauth2.TokenSource, string, error) {
	// Load credentials
	credBytes, err := os.ReadFile("configuration/credentials.json")
	if err != nil {
		return nil, "", err
	}
	config, err := google.ConfigFromJSON(credBytes, "https://mail.google.com/")
	if err != nil {
		return nil, "", err
	}

	// Load token
	tokenBytes, err := os.ReadFile("configuration/token.json")
	if err != nil {
		return nil, "", err
	}
	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, "", err
	}

	return config.TokenSource(context.Background(), &token), config.ClientID, nil
}

// XOAUTH2 Auth implementation
type xoauth2Auth struct {
	username    string
	accessToken string
	authString  string
}

func (a *xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(a.authString), nil
}
func (a *xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	return nil, nil
}

func initialise_oauth() error {
	tokenSource, _, err := getTokenSource()
	if err != nil {
		return fmt.Errorf("unable to get token source: %w", err)
	}
	OAUTH_TOKEN, err = tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}
	OAUTH_SET = true
	return nil
}
