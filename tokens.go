package main

import (
	"context"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

func PrintTokens(ctx context.Context) error {
	encoding := "cl100k_base"

	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return err
	}
	posts, err := q.GetItemsWithTitle(ctx, whoIsHiring)
	if err != nil {
		return err
	}
	for _, post := range posts {
		comments, err := q.GetItemsForParent(ctx, post.ID)
		if err != nil {
			return err
		}
		for _, c := range comments {
			// encode
			token := tke.Encode(c.Text, nil, nil)
			fmt.Println(c.ID, "has", len(token), "tokens")
		}
	}
	return nil
}
