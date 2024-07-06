package main

import (
	"context"
	"github.com/newhook/ai-hack/queries"
	"github.com/pkg/errors"
	"log/slog"
	"strconv"
)

func JobSearch(ctx context.Context, l *slog.Logger, months int, jobPrompt string) ([]queries.Item, []queries.Item, error) {
	terms, err := GetTerms(ctx, jobPrompt)
	if err != nil {
		return nil, nil, err
	}

	limit := 10
	queryResults, err := VectorSearch(ctx, l, months, embeddingModel, terms, limit)
	if err != nil {
		return nil, nil, err
	}

	type jobDescription struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}
	var descriptions []jobDescription
	for _, result := range queryResults {
		descriptions = append(descriptions, jobDescription{
			ID:      result.ID,
			Content: result.Item.Text,
		})
	}

	jobIDs, err := GetJobs(ctx, map[string]any{
		"Prompt": jobPrompt,
		"Jobs":   descriptions,
	})
	if err != nil {
		return nil, nil, err
	}

	// The comments must be contained in the original query results.
	var comments []queries.Item
	for _, id := range jobIDs {
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, nil, err
		}
		for _, result := range queryResults {
			if result.ID == n {
				comments = append(comments, result.Item)
				break
			}
		}
	}

	parentSet := NewSet[int]()
	for _, comment := range comments {
		parentSet.Add(comment.Parent)
	}
	allParents, err := q.GetItems(ctx, parentSet.Values())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	var parents []queries.Item
	for _, comment := range comments {
		for _, parent := range allParents {
			if comment.Parent == parent.ID {
				parents = append(parents, parent)
				break
			}
		}
	}
	if len(comments) != len(parents) {
		return nil, nil, errors.Errorf("mismatched comments and parents")
	}
	return comments, parents, nil
}
