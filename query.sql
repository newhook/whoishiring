-- name: InsertItem :exec
INSERT INTO items (
    id, deleted, type, by, time, text, dead, parent, poll, url, score, title, descendants
) VALUES (
     ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetItem :one
SELECT * from items where id = ?;

-- name: GetItemsWithTitle :many
select * from items where title like ? order by id desc;

-- name: GetItems :many
SELECT * from items where id in (sqlc.slice('ids'));

-- name: UpdateItem :exec
UPDATE items set parent = ?, time = ?, type = ?, by = ? where id = ?;

-- name: PaginateItems :many
SELECT * from items where id > ? order by id limit ?;

-- name: GetItemsForParent :many
SELECT * from items where parent = ? order by id;

-- name: GetPosts :many
select * from items where parent is null order by id desc;

-- name: GetKidsForItems :many
SELECT * from item_kids where item_id in (sqlc.slice('ids'));

-- name: GetEmbedding :one
select * from embeddings where item_id = ? and model = ?;

-- name: InsertEmbedding :exec
INSERT INTO embeddings(
    item_id, model, embedding, created_at, updated_at
) VALUES(
    ?, ?, ?, ?, ?
);

-- name: GetEmbeddingsByParent :many
select * from embeddings where model = ? and item_id in (select id from items where parent = ?);

-- name: GetEmbeddings :many
select * from embeddings where model = ? and item_id in (sqlc.slice('ids'));

-- name: InsertItemKids :exec
INSERT INTO item_kids (item_id, kid_id) VALUES (?, ?);

-- name: InsertItemParts :exec
INSERT INTO item_parts (item_id, part_id) VALUES (?, ?);
