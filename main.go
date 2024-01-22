package main

import (
	"fmt"
	"log"
	"os"
	"ufo_collector/authenticate"
	"ufo_collector/post"
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

		posts, nextAfter, err := post.GetRecent(token.AccessToken, after)
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
		id, secret, err = authenticate.GetCredsFromEnv("ANALYTICS_REDDIT_CLIENT_ID", "ANALYTICS_REDDIT_CLIENT_SECRET", "creds.env")
		if err != nil {
			log.Fatalf("Error reading credentials: %s", err)
		}
	} else if os.IsNotExist(err) {
		// If the file does not exist
		id, secret, err = authenticate.GetCredsFromEnv("ANALYTICS_REDDIT_CLIENT_ID", "ANALYTICS_REDDIT_CLIENT_SECRET")
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
