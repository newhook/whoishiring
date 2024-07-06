package main

import (
	"context"
	"github.com/lispad/go-generics-tools/binheap"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"sort"
	"sync"
	"time"
)

type Result struct {
	ID         int
	Term       string
	Similarity float32
	Item       queries.Item
}

func VectorSearch(ctx context.Context, l *slog.Logger, window int, model string, clause string, terms []string, limit int) ([]Result, error) {
	if window > MaxWindow {
		window = MaxWindow
	}

	posts, err := q.GetItemsWithTitle(ctx, clause)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	termVectors := make([][]float32, len(terms))
	for i, term := range terms {
		termVectors[i], err = GetEmbedding(ctx, term)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't create embedding of query")
		}
	}

	start := time.Now()
	results, err := searchPosts(ctx, limit, termVectors, posts[:window], model, terms)
	if err != nil {
		return nil, err
	}
	l.Info("results", slog.Int("results", len(results)), slog.Duration("in", time.Since(start)))

	results = deduplicateResults(results)
	l.Info("after deduplicating", slog.Int("results", len(results)))

	results, err = removeSimilarPosts(ctx, l, q, model, results)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	l.Info("after removing similar posts", slog.Int("results", len(results)))

	// Limit the results to the top N
	if len(results) > limit {
		results = results[:limit]
	}
	for _, result := range results {
		l.Info("result", slog.Int("id", result.ID), slog.Float64("similarity", float64(result.Similarity)), slog.String("term", result.Term))
	}
	return results, nil
}

func searchPosts(ctx context.Context, limit int, termVectors [][]float32, posts []queries.Item, model string, terms []string) ([]Result, error) {
	var mutex sync.Mutex
	h := binheap.EmptyTopNHeap[Result](limit*len(termVectors), func(i, j Result) bool {
		return i.Similarity > j.Similarity
	})

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)
	for _, post := range posts {
		g.Go(func() error {
			embeddings, err := q.GetEmbeddingsByParent(ctx, queries.GetEmbeddingsByParentParams{
				Model:  model,
				Parent: post.ID,
			})
			if err != nil {
				return errors.WithStack(err)
			}
			for _, embedding := range embeddings {
				if embedding.Embedding == nil {
					continue
				}
				ev, err := UnmarshalFloat32ArrayWithLength(embedding.Embedding)
				if err != nil {
					return errors.WithStack(err)
				}

				for i, termVector := range termVectors {
					sim, err := dotProduct(termVector, ev)
					if err != nil {
						return errors.WithStack(err)
					}
					mutex.Lock()
					h.Push(Result{
						ID:         embedding.ItemID,
						Term:       terms[i],
						Similarity: sim,
					})
					mutex.Unlock()
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return h.PopTopN(), nil
}

func dotProduct(a, b []float32) (float32, error) {
	// The vectors must have the same length
	if len(a) != len(b) {
		return 0, errors.New("vectors must have the same length")
	}

	var dotProduct float32
	for i := range a {
		dotProduct += a[i] * b[i]
	}

	return dotProduct, nil
}

func removeSimilarPosts(ctx context.Context, l *slog.Logger, q *queries.Queries, model string, results []Result) ([]Result, error) {
	var ids []int
	for _, r := range results {
		ids = append(ids, r.ID)
	}

	embeddings, err := q.GetEmbeddings(ctx, queries.GetEmbeddingsParams{
		Model: model,
		Ids:   ids,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	embeddingByID := map[int][]float32{}
	for _, e := range embeddings {
		v, err := UnmarshalFloat32ArrayWithLength(e.Embedding)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		embeddingByID[e.ItemID] = v
	}

	items, err := q.GetItems(ctx, ids)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	byPoster := map[string][]queries.Item{}
	for _, item := range items {
		byPoster[item.By] = append(byPoster[item.By], item)
	}

	itemByID := map[int]queries.Item{}
	for by, items := range byPoster {
		if len(items) > 1 {
			results, err := removeSimilarItems(items, embeddingByID)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			_ = by
			//l.Info("removed similar items", slog.String("by", by), slog.Int("before", len(items)), slog.Int("after", len(results)))
			for _, result := range results {
				itemByID[result.ID] = result
			}
		} else {
			for _, item := range items {
				itemByID[item.ID] = item
			}
		}
	}

	var matchedResults []Result
	for _, result := range results {
		if item, exists := itemByID[result.ID]; exists {
			result.Item = item
			matchedResults = append(matchedResults, result)
		}
	}
	sort.Slice(matchedResults, func(i, j int) bool {
		return matchedResults[i].Similarity > matchedResults[j].Similarity
	})

	return matchedResults, nil
}

func removeSimilarItems(items []queries.Item, embeddings map[int][]float32) ([]queries.Item, error) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			sim, err := dotProduct(embeddings[items[i].ID], embeddings[items[j].ID])
			if err != nil {
				return nil, errors.WithStack(err)
			}
			if sim > 0.9 {
				items = append(items[:j], items[j+1:]...)
				j--
			}
		}
	}
	return items, nil
}

func deduplicateResults(results []Result) []Result {
	dedup := NewSet[int]()
	for i := 0; i < len(results); i++ {
		id := results[i].ID
		if dedup.Contains(id) {
			results = append(results[:i], results[i+1:]...)
		} else {
			dedup.Add(id)
		}
	}
	return results
}
