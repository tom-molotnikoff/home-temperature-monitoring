package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	b, err := os.ReadFile("../configuration/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://mail.google.com/")
	if err != nil {
		log.Fatalf("Unable to parse credentials.json: %v", err)
	}

	// Use the first redirect URI from credentials.json
	redirectURL := config.RedirectURL
	if redirectURL == "" {
		redirectURL = "http://localhost:8080"
	}
	config.RedirectURL = redirectURL

	// Start local server to receive the code
	codeCh := make(chan string)
	srv := &http.Server{Addr: ":8080"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authorization received. You can close this window.")
			codeCh <- code
		} else {
			http.Error(w, "No code in request", http.StatusBadRequest)
		}
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	code := <-codeCh
	srv.Shutdown(context.Background())

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	f, err := os.Create("../configuration/token.json")
	if err != nil {
		log.Fatalf("Unable to create token.json: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	fmt.Println("Token saved to ../configuration/token.json.")
}
