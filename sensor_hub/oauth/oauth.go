package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

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

func startOAuthTokenRefresher() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			tokenSource, _, err := getTokenSource()
			if err != nil {
				fmt.Printf("OAuth: unable to get token source: %v\n", err)
				continue
			}
			token, err := tokenSource.Token()
			if err != nil {
				fmt.Printf("OAuth: unable to refresh token: %v\n", err)
			} else {
				OAUTH_TOKEN = token
				OAUTH_SET = true
				tokenBytes, err := json.Marshal(token)
				if err != nil {
					fmt.Printf("OAuth: unable to marshal token: %v\n", err)
					continue
				}
				err = os.WriteFile("configuration/token.json", tokenBytes, 0600)
				if err != nil {
					fmt.Printf("OAuth: unable to write token.json: %v\n", err)
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
	OAUTH_TOKEN, err = tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}
	startOAuthTokenRefresher()
	OAUTH_SET = true
	return nil
}
