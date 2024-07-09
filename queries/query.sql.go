// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package queries

import (
	"context"
	"strings"
)

const getEmbedding = `-- name: GetEmbedding :one
select id, model, item_id, embedding, created_at, updated_at from embeddings where item_id = ? and model = ?
`

type GetEmbeddingParams struct {
	ItemID int    `json:"item_id"`
	Model  string `json:"model"`
}

func (q *Queries) GetEmbedding(ctx context.Context, arg GetEmbeddingParams) (Embedding, error) {
	row := q.db.QueryRowContext(ctx, getEmbedding, arg.ItemID, arg.Model)
	var i Embedding
	err := row.Scan(
		&i.ID,
		&i.Model,
		&i.ItemID,
		&i.Embedding,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getEmbeddings = `-- name: GetEmbeddings :many
select id, model, item_id, embedding, created_at, updated_at from embeddings where model = ? and item_id in (/*SLICE:ids*/?)
`

type GetEmbeddingsParams struct {
	Model string `json:"model"`
	Ids   []int  `json:"ids"`
}

func (q *Queries) GetEmbeddings(ctx context.Context, arg GetEmbeddingsParams) ([]Embedding, error) {
	query := getEmbeddings
	var queryParams []interface{}
	queryParams = append(queryParams, arg.Model)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Embedding
	for rows.Next() {
		var i Embedding
		if err := rows.Scan(
			&i.ID,
			&i.Model,
			&i.ItemID,
			&i.Embedding,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEmbeddingsByParent = `-- name: GetEmbeddingsByParent :many
select id, model, item_id, embedding, created_at, updated_at from embeddings where model = ? and item_id in (select id from items where parent = ?)
`

type GetEmbeddingsByParentParams struct {
	Model  string `json:"model"`
	Parent int    `json:"parent"`
}

func (q *Queries) GetEmbeddingsByParent(ctx context.Context, arg GetEmbeddingsByParentParams) ([]Embedding, error) {
	rows, err := q.db.QueryContext(ctx, getEmbeddingsByParent, arg.Model, arg.Parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Embedding
	for rows.Next() {
		var i Embedding
		if err := rows.Scan(
			&i.ID,
			&i.Model,
			&i.ItemID,
			&i.Embedding,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getItem = `-- name: GetItem :one
SELECT id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where id = ?
`

func (q *Queries) GetItem(ctx context.Context, id int) (Item, error) {
	row := q.db.QueryRowContext(ctx, getItem, id)
	var i Item
	err := row.Scan(
		&i.ID,
		&i.Deleted,
		&i.Type,
		&i.By,
		&i.Time,
		&i.Text,
		&i.Dead,
		&i.Parent,
		&i.Poll,
		&i.Url,
		&i.Score,
		&i.Title,
		&i.Descendants,
	)
	return i, err
}

const getItems = `-- name: GetItems :many
SELECT id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where id in (/*SLICE:ids*/?)
`

func (q *Queries) GetItems(ctx context.Context, ids []int) ([]Item, error) {
	query := getItems
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Type,
			&i.By,
			&i.Time,
			&i.Text,
			&i.Dead,
			&i.Parent,
			&i.Poll,
			&i.Url,
			&i.Score,
			&i.Title,
			&i.Descendants,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getItemsForParent = `-- name: GetItemsForParent :many
SELECT id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where parent = ? order by id
`

func (q *Queries) GetItemsForParent(ctx context.Context, parent int) ([]Item, error) {
	rows, err := q.db.QueryContext(ctx, getItemsForParent, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Type,
			&i.By,
			&i.Time,
			&i.Text,
			&i.Dead,
			&i.Parent,
			&i.Poll,
			&i.Url,
			&i.Score,
			&i.Title,
			&i.Descendants,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getItemsWithTitle = `-- name: GetItemsWithTitle :many
select id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where title like ? order by id desc
`

func (q *Queries) GetItemsWithTitle(ctx context.Context, title string) ([]Item, error) {
	rows, err := q.db.QueryContext(ctx, getItemsWithTitle, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Type,
			&i.By,
			&i.Time,
			&i.Text,
			&i.Dead,
			&i.Parent,
			&i.Poll,
			&i.Url,
			&i.Score,
			&i.Title,
			&i.Descendants,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getKidsForItems = `-- name: GetKidsForItems :many
SELECT item_id, kid_id from item_kids where item_id in (/*SLICE:ids*/?)
`

func (q *Queries) GetKidsForItems(ctx context.Context, ids []int) ([]ItemKid, error) {
	query := getKidsForItems
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ItemKid
	for rows.Next() {
		var i ItemKid
		if err := rows.Scan(&i.ItemID, &i.KidID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLinkedInScrape = `-- name: GetLinkedInScrape :one
select url, json, created_at, updated_at from linkedin_scrapes where url = ?
`

func (q *Queries) GetLinkedInScrape(ctx context.Context, url string) (LinkedinScrape, error) {
	row := q.db.QueryRowContext(ctx, getLinkedInScrape, url)
	var i LinkedinScrape
	err := row.Scan(
		&i.Url,
		&i.Json,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPosts = `-- name: GetPosts :many
select id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where parent is null order by id desc
`

func (q *Queries) GetPosts(ctx context.Context) ([]Item, error) {
	rows, err := q.db.QueryContext(ctx, getPosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Type,
			&i.By,
			&i.Time,
			&i.Text,
			&i.Dead,
			&i.Parent,
			&i.Poll,
			&i.Url,
			&i.Score,
			&i.Title,
			&i.Descendants,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertEmbedding = `-- name: InsertEmbedding :exec
INSERT INTO embeddings(
    item_id, model, embedding, created_at, updated_at
) VALUES(
    ?, ?, ?, ?, ?
)
`

type InsertEmbeddingParams struct {
	ItemID    int    `json:"item_id"`
	Model     string `json:"model"`
	Embedding []byte `json:"embedding"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

func (q *Queries) InsertEmbedding(ctx context.Context, arg InsertEmbeddingParams) error {
	_, err := q.db.ExecContext(ctx, insertEmbedding,
		arg.ItemID,
		arg.Model,
		arg.Embedding,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const insertItem = `-- name: InsertItem :exec
INSERT INTO items (
    id, deleted, type, by, time, text, dead, parent, poll, url, score, title, descendants
) VALUES (
     ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
`

type InsertItemParams struct {
	ID          int    `json:"id"`
	Deleted     bool   `json:"deleted"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int    `json:"time"`
	Text        string `json:"text"`
	Dead        bool   `json:"dead"`
	Parent      int    `json:"parent"`
	Poll        int    `json:"poll"`
	Url         string `json:"url"`
	Score       int    `json:"score"`
	Title       string `json:"title"`
	Descendants int    `json:"descendants"`
}

func (q *Queries) InsertItem(ctx context.Context, arg InsertItemParams) error {
	_, err := q.db.ExecContext(ctx, insertItem,
		arg.ID,
		arg.Deleted,
		arg.Type,
		arg.By,
		arg.Time,
		arg.Text,
		arg.Dead,
		arg.Parent,
		arg.Poll,
		arg.Url,
		arg.Score,
		arg.Title,
		arg.Descendants,
	)
	return err
}

const insertItemKids = `-- name: InsertItemKids :exec
INSERT INTO item_kids (item_id, kid_id) VALUES (?, ?)
`

type InsertItemKidsParams struct {
	ItemID int `json:"item_id"`
	KidID  int `json:"kid_id"`
}

func (q *Queries) InsertItemKids(ctx context.Context, arg InsertItemKidsParams) error {
	_, err := q.db.ExecContext(ctx, insertItemKids, arg.ItemID, arg.KidID)
	return err
}

const insertItemParts = `-- name: InsertItemParts :exec
INSERT INTO item_parts (item_id, part_id) VALUES (?, ?)
`

type InsertItemPartsParams struct {
	ItemID int `json:"item_id"`
	PartID int `json:"part_id"`
}

func (q *Queries) InsertItemParts(ctx context.Context, arg InsertItemPartsParams) error {
	_, err := q.db.ExecContext(ctx, insertItemParts, arg.ItemID, arg.PartID)
	return err
}

const insertLinkedInScrape = `-- name: InsertLinkedInScrape :exec
INSERT INTO linkedin_scrapes (url, json, created_at, updated_at) VALUES (?, ?, ?, ?)
`

type InsertLinkedInScrapeParams struct {
	Url       string `json:"url"`
	Json      string `json:"json"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

func (q *Queries) InsertLinkedInScrape(ctx context.Context, arg InsertLinkedInScrapeParams) error {
	_, err := q.db.ExecContext(ctx, insertLinkedInScrape,
		arg.Url,
		arg.Json,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const paginateItems = `-- name: PaginateItems :many
SELECT id, deleted, type, "by", time, text, dead, parent, poll, url, score, title, descendants from items where id > ? order by id limit ?
`

type PaginateItemsParams struct {
	ID    int   `json:"id"`
	Limit int64 `json:"limit"`
}

func (q *Queries) PaginateItems(ctx context.Context, arg PaginateItemsParams) ([]Item, error) {
	rows, err := q.db.QueryContext(ctx, paginateItems, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(
			&i.ID,
			&i.Deleted,
			&i.Type,
			&i.By,
			&i.Time,
			&i.Text,
			&i.Dead,
			&i.Parent,
			&i.Poll,
			&i.Url,
			&i.Score,
			&i.Title,
			&i.Descendants,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateItem = `-- name: UpdateItem :exec
UPDATE items set parent = ?, time = ?, type = ?, by = ? where id = ?
`

type UpdateItemParams struct {
	Parent int    `json:"parent"`
	Time   int    `json:"time"`
	Type   string `json:"type"`
	By     string `json:"by"`
	ID     int    `json:"id"`
}

func (q *Queries) UpdateItem(ctx context.Context, arg UpdateItemParams) error {
	_, err := q.db.ExecContext(ctx, updateItem,
		arg.Parent,
		arg.Time,
		arg.Type,
		arg.By,
		arg.ID,
	)
	return err
}
