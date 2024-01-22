package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"ufo_collector/authenticate"
)

func main() {
	token, err := login()
	if err != nil {
		log.Fatalf("Error logging in: %s", err)
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

func login() (*authenticate.TokenResponse, error) {
	var id, secret string
	var err error

	if _, err := os.Stat("creds.env"); err == nil {
		// If the file exists
		id, secret, err = authenticate.GetCredsFromEnv("ANALYTICS_CLIENT_ID", "ANALYTICS_CLIENT_SECRET", "creds.env")
		if err != nil {
			log.Fatalf("Error reading credentials: %s", err)
		}
	} else if os.IsNotExist(err) {
		// If the file does not exist
		id, secret, err = authenticate.GetCredsFromEnv("ANALYTICS_CLIENT_ID", "ANALYTICS_CLIENT_SECRET")
		if err != nil {
			log.Fatalf("Error reading credentials: %s", err)
		}
	} else {
		// If there's an error checking the file
		log.Fatalf("Error checking creds.env file: %s", err)
	}

	token, err := authenticate.GetToken(id, secret)
	if err != nil {
		log.Fatalf("Error getting token: %s", err)
	}

	return token, nil
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
