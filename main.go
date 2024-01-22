package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"gopkg.in/yaml.v2"
)

type AppCreds struct {
	ClientID     string `yaml:"ClientID"`
	ClientSecret string `yaml:"ClientSecret"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func main() {
	creds, err := getCredsFromEnvOrFile("creds.yaml")
	if err != nil {
		fmt.Println("Error reading credentials:", err)
		return
	}

	token, err := getAppToken(creds)
	if err != nil {
		fmt.Println("Error getting app token:", err)
		return
	}

	ids := make(map[string]bool)
	uniqueFlairs := make(map[string]bool)
	totalComments := 0
	after := ""
	for i := 0; i < 10; i++ { // get 5 pages of posts
		posts, nextAfter, err := getRecentPosts(token.AccessToken, after)
		if err != nil {
			fmt.Println("Error getting recent posts:", err)
			return
		}

		fmt.Printf("Processing page %d, %d posts\n", i+1, len(posts))

		for _, post := range posts {
			ids[post.ID] = true
			uniqueFlairs[post.Flair] = true
			totalComments += post.NumComments
		}

		after = nextAfter
	}

	fmt.Println("Total posts:", len(ids))

	fmt.Println("Total comments:", totalComments)

	fmt.Println("Unique flairs:")
	for flair := range uniqueFlairs {
		fmt.Println(flair)
	}
}

func getCredsFromEnvOrFile(filename string) (AppCreds, error) {
	clientID := os.Getenv("ANALYTICS_CLIENT_ID")
	clientSecret := os.Getenv("ANALYTICS_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return AppCreds{ClientID: clientID, ClientSecret: clientSecret}, nil
	}

	return getCredsFromFile(filename)
}

func getCredsFromFile(filename string) (AppCreds, error) {
	var creds AppCreds

	data, err := os.ReadFile(filename)
	if err != nil {
		return creds, err
	}

	err = yaml.Unmarshal(data, &creds)
	if err != nil {
		return creds, err
	}

	return creds, nil
}

func getAppToken(creds AppCreds) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
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

type Post struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Flair       string `json:"link_flair_text"`
	NumComments int    `json:"num_comments"`
}

type PostListing struct {
	Data struct {
		Children []struct {
			Data Post `json:"data"`
		} `json:"children"`
		After string `json:"after"`
	} `json:"data"`
}

func getRecentPosts(token string, after string) ([]Post, string, error) {
	url := "https://oauth.reddit.com/r/UFOs/new?limit=100"
	if after != "" {
		url += "&after=" + after
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("User-Agent", "MyApp/0.0.1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var postListing PostListing
	err = json.Unmarshal(body, &postListing)
	if err != nil {
		return nil, "", err
	}

	posts := make([]Post, len(postListing.Data.Children))
	for i, child := range postListing.Data.Children {
		posts[i] = child.Data
	}

	return posts, postListing.Data.After, nil
}
