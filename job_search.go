package main

import (
	"context"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	"log/slog"
	"strconv"
	"time"
)

func JobSearch(ctx context.Context, l *slog.Logger, months int, searchType string, jobPrompt string) ([]queries.Item, []queries.Item, error) {
	// search type
	// - hiring: who is hiring
	// - seekers: who wants to be hired
	var clause string
	if searchType == "hiring" {
		clause = whoIsHiring
	} else if searchType == "seekers" {
		clause = whoWantsToBeHired
	} else {
		return nil, nil, errors.Errorf("invalid search type: %s", searchType)
	}

	terms, err := GetTerms(ctx, jobPrompt)
	if err != nil {
		return nil, nil, err
	}

	limit := 10
	queryResults, err := VectorSearch(ctx, l, months, *embeddingModel, clause, terms, limit)
	if err != nil {
		return nil, nil, err
	}

	type jobDescription struct {
		ID      int    `json:"id"`
		Date    string `json:"date"`
		Content string `json:"content"`
	}
	var descriptions []jobDescription
	for _, result := range queryResults {
		descriptions = append(descriptions, jobDescription{
			ID:      result.ID,
			Date:    time.Unix(int64(result.Item.Time), 0).String(),
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
