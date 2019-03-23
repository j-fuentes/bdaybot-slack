package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func GetHTTPClient(config *oauth2.Config, token *oauth2.Token) *http.Client {
	return config.Client(context.Background(), token)
}

func ReadTokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func RetrieveTokenInteractively(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code:\n\n%v\n", authURL)

	fmt.Println("\nPaste here the authorization code:")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}

	return token, nil
}

func WriteTokenToFile(token *oauth2.Token, path string) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}
