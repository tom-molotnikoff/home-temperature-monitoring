package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OauthToken *oauth2.Token
var OauthSet = false

type XOauth2Auth struct {
	Username    string
	AccessToken string
	AuthString  string
}

func (a *XOauth2Auth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(a.AuthString), nil
}
func (a *XOauth2Auth) Next(_ []byte, _ bool) ([]byte, error) {
	return nil, nil
}

func getTokenSource() (oauth2.TokenSource, string, error) {
	// Load credentials
	credBytes, err := os.ReadFile("configuration/credentials.json")
	if err != nil {
		return nil, "", fmt.Errorf("unable to read credentials.json: %w", err)
	}
	config, err := google.ConfigFromJSON(credBytes, "https://mail.google.com/")
	if err != nil {
		return nil, "", fmt.Errorf("unable to parse credentials.json: %w", err)
	}

	// Load token
	tokenBytes, err := os.ReadFile("configuration/token.json")
	if err != nil {
		return nil, "", fmt.Errorf("unable to read token.json: %w", err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, "", fmt.Errorf("unable to unmarshal token.json: %w", err)
	}

	return config.TokenSource(context.Background(), &token), config.ClientID, nil
}

func startOAuthTokenRefresher() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			tokenSource, _, err := getTokenSource()
			if err != nil {
				fmt.Printf("OAuth: unable to get token source: %v", err)
				continue
			}
			token, err := tokenSource.Token()
			if err != nil {
				fmt.Printf("OAuth: unable to refresh token: %v", err)
			} else {
				OauthToken = token
				OauthSet = true
				tokenBytes, err := json.Marshal(token)
				if err != nil {
					fmt.Printf("OAuth: unable to marshal token: %v", err)
					continue
				}
				err = os.WriteFile("configuration/token.json", tokenBytes, 0600)
				if err != nil {
					fmt.Printf("OAuth: unable to write token.json: %v", err)
				}
			}
		}
	}()
}

func InitialiseOauth() error {
	tokenSource, _, err := getTokenSource()
	if err != nil {
		return fmt.Errorf("unable to get token source: %w", err)
	}
	OauthToken, err = tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}
	startOAuthTokenRefresher()
	OauthSet = true
	return nil
}
