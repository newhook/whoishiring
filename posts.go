package main

import (
	"context"
	"github.com/newhook/whoishiring/hn"
	"github.com/newhook/whoishiring/queries"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
)

func FetchPosts(ctx context.Context, l *slog.Logger, q *queries.Queries) error {
	u, err := hn.GetUser(ctx, "whoishiring")
	if err != nil {
		return errors.WithStack(err)
	}

	start := time.Now()
	submissions, err := fetchPostsById(ctx, l, q, u.Submitted)
	if err != nil {
		return err
	}
	l.Info("fetched submissions", slog.Int("count", len(submissions)),
		slog.Duration("elapsed", time.Since(start)))

	kids, err := q.GetKidsForItems(ctx, u.Submitted)
	if err != nil {
		return errors.WithStack(err)
	}

	var ids []int
	for _, kid := range kids {
		ids = append(ids, kid.KidID)
	}

	start = time.Now()
	items, err := fetchPostsById(ctx, l, q, ids)
	if err != nil {
		return err
	}
	l.Info("fetched children", slog.Int("count", len(items)),
		slog.Duration("elapsed", time.Since(start)))
	return nil
}

func fetchPostsById(ctx context.Context, l *slog.Logger, q *queries.Queries, itemIDs []int) ([]queries.Item, error) {
	toDownload := NewSet[int](itemIDs...)
	items, err := q.GetItemsBatch(ctx, itemIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, item := range items {
		toDownload.Remove(item.ID)
	}

	if err := downloadStoreItems(ctx, l, q, toDownload.Values()); err != nil {
		return nil, err
	}

	downloaded, err := q.GetItemsBatch(ctx, toDownload.Values())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return append(items, downloaded...), nil
}

func downloadStoreItems(ctx context.Context, l *slog.Logger, q *queries.Queries, toFetch []int) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)
	max := len(toFetch)
	l.Info("downloading and inserting items", slog.Int("count", max))
	for _, id := range toFetch {
		g.Go(func() error {
			//fmt.Printf("%05d to fetch out of %d\r", len(toFetch), max)
			item, err := hn.GetItem(ctx, id)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch item %d", id)
			}
			return insertItem(ctx, q, item)
		})
	}
	return g.Wait()
}

func insertItem(ctx context.Context, q *queries.Queries, item hn.Item) error {
	err := q.InsertItem(ctx, queries.InsertItemParams{
		ID:          item.ID,
		Deleted:     item.Deleted,
		Type:        item.Type,
		By:          item.By,
		Time:        item.Time,
		Text:        item.Text,
		Dead:        item.Dead,
		Parent:      item.Parent,
		Poll:        item.Poll,
		Url:         item.URL,
		Score:       item.Score,
		Title:       item.Title,
		Descendants: item.Descendants,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	for _, kid := range item.Kids {
		err = q.InsertItemKids(ctx, queries.InsertItemKidsParams{
			ItemID: item.ID,
			KidID:  kid,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	for _, part := range item.Parts {
		err = q.InsertItemParts(ctx, queries.InsertItemPartsParams{
			ItemID: item.ID,
			PartID: part,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
