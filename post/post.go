package post

import (
	"encoding/json"
	"io"
	"net/http"
)

type Entity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Flair       string `json:"link_flair_text"`
	NumComments int    `json:"num_comments"`
	Author      string `json:"author"` // Add this line
	Created     int    `json:"created_utc"`
}

type Listing struct {
	Data struct {
		Children []struct {
			Data Entity `json:"data"`
		} `json:"children"`
		After string `json:"after"`
	} `json:"data"`
}

func GetRecent(token string, after string) ([]Entity, string, error) {
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

	var postListing Listing
	err = json.Unmarshal(body, &postListing)
	if err != nil {
		return nil, "", err
	}

	posts := make([]Entity, len(postListing.Data.Children))
	for i, child := range postListing.Data.Children {
		posts[i] = child.Data
	}

	return posts, postListing.Data.After, nil
}
