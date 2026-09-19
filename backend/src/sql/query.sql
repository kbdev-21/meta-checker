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

-- Rank solo/flex truyền vào NULL (unknown) thì giữ nguyên cả nhóm cột rank cũ.
-- name: UpsertPlayer :exec
INSERT INTO players (
    id, server, name, tag, normalized_name, normalized_tag, profile_icon_id, summoner_level, search_string,
    solo_rank, solo_tier, solo_lp, solo_rank_power, solo_wins, solo_losses,
    flex_rank, flex_tier, flex_lp, flex_wins, flex_losses
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
ON CONFLICT (id) DO UPDATE SET
    server          = EXCLUDED.server,
    name            = EXCLUDED.name,
    tag             = EXCLUDED.tag,
    normalized_name = EXCLUDED.normalized_name,
    normalized_tag  = EXCLUDED.normalized_tag,
    profile_icon_id = COALESCE(EXCLUDED.profile_icon_id, players.profile_icon_id),
    summoner_level  = COALESCE(EXCLUDED.summoner_level, players.summoner_level),
    search_string   = EXCLUDED.search_string,

    solo_rank       = COALESCE(EXCLUDED.solo_rank, players.solo_rank),
    solo_tier       = CASE WHEN EXCLUDED.solo_rank IS NULL THEN players.solo_tier       ELSE EXCLUDED.solo_tier       END,
    solo_lp         = CASE WHEN EXCLUDED.solo_rank IS NULL THEN players.solo_lp         ELSE EXCLUDED.solo_lp         END,
    solo_rank_power = CASE WHEN EXCLUDED.solo_rank IS NULL THEN players.solo_rank_power ELSE EXCLUDED.solo_rank_power END,
    solo_wins       = CASE WHEN EXCLUDED.solo_rank IS NULL THEN players.solo_wins       ELSE EXCLUDED.solo_wins       END,
    solo_losses     = CASE WHEN EXCLUDED.solo_rank IS NULL THEN players.solo_losses     ELSE EXCLUDED.solo_losses     END,

    flex_rank       = COALESCE(EXCLUDED.flex_rank, players.flex_rank),
    flex_tier       = CASE WHEN EXCLUDED.flex_rank IS NULL THEN players.flex_tier   ELSE EXCLUDED.flex_tier   END,
    flex_lp         = CASE WHEN EXCLUDED.flex_rank IS NULL THEN players.flex_lp     ELSE EXCLUDED.flex_lp     END,
    flex_wins       = CASE WHEN EXCLUDED.flex_rank IS NULL THEN players.flex_wins   ELSE EXCLUDED.flex_wins   END,
    flex_losses     = CASE WHEN EXCLUDED.flex_rank IS NULL THEN players.flex_losses ELSE EXCLUDED.flex_losses END,

    updated_at      = now();

-- name: GetPlayerById :one
SELECT * FROM players WHERE id = $1;

-- Truyền vào name / tag đã normalize (chữ thường + trim).
-- name: GetPlayerByNameAndTag :one
SELECT * FROM players WHERE normalized_name = sqlc.arg(normalized_name) AND normalized_tag = sqlc.arg(normalized_tag);

-- name: SearchPlayers :many
SELECT * FROM players
WHERE name ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR tag ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR search_string LIKE '%' || sqlc.arg(normalized_keyword)::text || '%'
ORDER BY solo_rank_power DESC NULLS LAST
LIMIT sqlc.arg(lim);
