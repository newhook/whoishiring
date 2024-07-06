package queries

import "context"

func (q *Queries) GetItemsBatch(ctx context.Context, ids []int) ([]Item, error) {
	max := 500
	var items []Item
	query := func(ids []int) error {
		queried, err := q.GetItems(ctx, ids)
		if err != nil {
			return err
		}
		items = append(items, queried...)
		return nil
	}
	for len(ids) > max {
		if err := query(ids[:max]); err != nil {
			return nil, err
		}
		ids = ids[max:]
	}
	if err := query(ids); err != nil {
		return nil, err
	}
	return items, nil
}
