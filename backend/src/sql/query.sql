-- name: UpsertChampion :exec
INSERT INTO champions (id, slug, name, title, img_url, version)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    slug       = EXCLUDED.slug,
    name       = EXCLUDED.name,
    title      = EXCLUDED.title,
    img_url    = EXCLUDED.img_url,
    version    = EXCLUDED.version,
    updated_at = now();

-- name: UpsertItem :exec
INSERT INTO items (id, name, plaintext, type, gold_total, from_items, into_items, is_sr, img_url, version)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name       = EXCLUDED.name,
    plaintext  = EXCLUDED.plaintext,
    type       = EXCLUDED.type,
    gold_total = EXCLUDED.gold_total,
    from_items = EXCLUDED.from_items,
    into_items = EXCLUDED.into_items,
    is_sr      = EXCLUDED.is_sr,
    img_url    = EXCLUDED.img_url,
    version    = EXCLUDED.version,
    updated_at = now();

-- name: ListChampions :many
SELECT * FROM champions ORDER BY name;

-- name: ListItems :many
SELECT * FROM items ORDER BY id;
