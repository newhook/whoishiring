package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Item struct {
	ID          int    `json:"id"`
	Deleted     bool   `json:"deleted,omitempty"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int    `json:"time"`
	Text        string `json:"text,omitempty"`
	Dead        bool   `json:"dead,omitempty"`
	Parent      int    `json:"parent,omitempty"`
	Poll        int    `json:"poll,omitempty"`
	Kids        []int  `json:"kids,omitempty"`
	URL         string `json:"url,omitempty"`
	Score       int    `json:"score,omitempty"`
	Title       string `json:"title,omitempty"`
	Parts       []int  `json:"parts,omitempty"`
	Descendants int    `json:"descendants,omitempty"`
}

type User struct {
	About     string `json:"about"`
	Created   int    `json:"created"`
	ID        string `json:"id"`
	Karma     int    `json:"karma"`
	Submitted []int  `json:"submitted"`
}

func GetUser(ctx context.Context, user string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://hacker-news.firebaseio.com/v0/user/%s.json", user), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var u User
	err = json.Unmarshal(body, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func GetItem(ctx context.Context, id int) (Item, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id), nil)
	if err != nil {
		return Item{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Item{}, err
	}
	if err != nil {
		return Item{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Item{}, err
	}

	var item Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		return Item{}, err
	}

	return item, nil
}
