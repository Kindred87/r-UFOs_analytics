package authenticate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

// Creds holds the client ID and secret.
type Creds struct {
	ClientID     string `yaml:"ClientID"`
	ClientSecret string `yaml:"ClientSecret"`
}

// TokenResponse holds the response from the token request.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// GetToken returns a token for the given client ID and secret.
func GetToken(id, secret string) (*TokenResponse, error) {

	token, err := getAppToken(id, secret)
	if err != nil {
		fmt.Println("Error getting app token:", err)
		return nil, fmt.Errorf("error getting app token: %s", err)
	}

	return token, nil
}

// GetCredsFromEnv returns credentials from the given environment variables.
func GetCredsFromEnv(idKey, secretKey string, files ...string) (string, string, error) {
	if len(files) > 0 {
		err := godotenv.Load(files...)
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
			return "", "", fmt.Errorf("error loading .env file: %s", err)
		}
	}

	clientID := os.Getenv(idKey)
	clientSecret := os.Getenv(secretKey)

	if clientID == "" && clientSecret == "" {
		return "", "", fmt.Errorf("could not find credentials in environment variables %s and %s", idKey, secretKey)
	}

	return clientID, clientSecret, nil
}

func getAppToken(id, secret string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(id, secret)
	req.Header.Add("User-Agent", "MyApp/0.0.1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}
