package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"ufo_collector/authenticate"
	"ufo_collector/database"
	"ufo_collector/post"
)

func main() {
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %s", err)
	}
	defer database.CloseDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Every ten minutes, sync the database
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := database.SyncDB()
			if err != nil {
				fmt.Println("Error syncing database:", err)
			}
			fmt.Println("Synced database")
			<-time.After(10 * time.Minute)
		}
	}()

	token, err := login()
	if err != nil {
		log.Fatalf("Error logging in: %s", err)
	}

	after := ""
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var posts []post.Entity
		posts, after, err = post.GetRecent(token.AccessToken, after)
		if err != nil {
			fmt.Println("Error getting recent posts:", err)
			continue
		}

		log.Printf("Retrieved %d posts at %s\n", len(posts), time.Now().Format("01/02/2006 15:04:05 PST"))

		for _, post := range posts {
			// Convert UTC time to local time
			t := time.Unix(int64(post.Created), 0)
			// Format the time as mm/dd/yyyy hh:mm:ss pacific time
			timestamp := t.Format("01/02/2006 15:04:05 PST")
			err = database.AddPostHistory(post.ID, timestamp, post.Flair, post.URL, post.Author, post.NumComments)
			if err != nil {
				fmt.Println("Error adding post history for post", post.ID, ":", err)
			}
		}

		// If there are 100 posts, get the next page immediately
		if len(posts) != 100 {
			<-time.After(30 * time.Second)
		}
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
