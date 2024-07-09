CREATE TABLE IF NOT EXISTS items (
    id INT PRIMARY KEY NOT NULL,
    deleted TINYINT NOT NULL,
    type text NOT NULL,
    by text NOT NULL,
    time INT NOT NULL,
    text TEXT NOT NULL,
    dead TINYINT NOT NULL,
    parent INT NOT NULL,
    poll INT NOT NULL,
    url TEXT NOT NULL,
    score INT NOT NULL,
    title TEXT NOT NULL,
    descendants INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_parent_id ON items(parent);

CREATE TABLE IF NOT EXISTS embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    model TEXT NOT NULL,
    item_id INTEGER NOT NULL,
    embedding BLOB NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_embeddings_model_item_id ON embeddings(model, item_id);

CREATE TABLE IF NOT EXISTS item_kids (
    item_id INT NOT NULL,
    kid_id INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_item_kids_item_id ON item_kids(item_id);

CREATE TABLE IF NOT EXISTS item_parts (
    item_id INT NOT NULL,
    part_id INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_item_parents_item_id ON item_parts(item_id);

CREATE TABLE IF NOT EXISTS linkedin_scrapes (
    url text PRIMARY KEY NOT NULL,
    json text NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);