-- name: UpsertChampion :exec
INSERT INTO lol_champions (id, slug, name, title, img_url, patch)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    slug       = EXCLUDED.slug,
    name       = EXCLUDED.name,
    title      = EXCLUDED.title,
    img_url    = EXCLUDED.img_url,
    patch      = EXCLUDED.patch,
    updated_at = now();

-- name: UpsertItem :exec
INSERT INTO lol_items (id, name, plaintext, type, gold_total, from_items, into_items, is_sr, img_url, patch)
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
    patch      = EXCLUDED.patch,
    updated_at = now();

-- name: ListChampions :many
SELECT * FROM lol_champions ORDER BY name;

-- name: ListItems :many
SELECT * FROM lol_items ORDER BY id;

-- Rank solo/flex truyền vào NULL (unknown) thì giữ nguyên cả nhóm cột rank cũ.
-- name: UpsertPlayer :exec
INSERT INTO lol_players (
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
    profile_icon_id = COALESCE(EXCLUDED.profile_icon_id, lol_players.profile_icon_id),
    summoner_level  = COALESCE(EXCLUDED.summoner_level, lol_players.summoner_level),
    search_string   = EXCLUDED.search_string,

    solo_rank       = COALESCE(EXCLUDED.solo_rank, lol_players.solo_rank),
    solo_tier       = CASE WHEN EXCLUDED.solo_rank IS NULL THEN lol_players.solo_tier       ELSE EXCLUDED.solo_tier       END,
    solo_lp         = CASE WHEN EXCLUDED.solo_rank IS NULL THEN lol_players.solo_lp         ELSE EXCLUDED.solo_lp         END,
    solo_rank_power = CASE WHEN EXCLUDED.solo_rank IS NULL THEN lol_players.solo_rank_power ELSE EXCLUDED.solo_rank_power END,
    solo_wins       = CASE WHEN EXCLUDED.solo_rank IS NULL THEN lol_players.solo_wins       ELSE EXCLUDED.solo_wins       END,
    solo_losses     = CASE WHEN EXCLUDED.solo_rank IS NULL THEN lol_players.solo_losses     ELSE EXCLUDED.solo_losses     END,

    flex_rank       = COALESCE(EXCLUDED.flex_rank, lol_players.flex_rank),
    flex_tier       = CASE WHEN EXCLUDED.flex_rank IS NULL THEN lol_players.flex_tier   ELSE EXCLUDED.flex_tier   END,
    flex_lp         = CASE WHEN EXCLUDED.flex_rank IS NULL THEN lol_players.flex_lp     ELSE EXCLUDED.flex_lp     END,
    flex_wins       = CASE WHEN EXCLUDED.flex_rank IS NULL THEN lol_players.flex_wins   ELSE EXCLUDED.flex_wins   END,
    flex_losses     = CASE WHEN EXCLUDED.flex_rank IS NULL THEN lol_players.flex_losses ELSE EXCLUDED.flex_losses END,

    updated_at      = now();

-- name: GetPlayerById :one
SELECT * FROM lol_players WHERE id = $1;

-- Id không có trong DB thì bỏ qua.
-- name: GetPlayersByIds :many
SELECT * FROM lol_players WHERE id = ANY(sqlc.arg(ids)::text[]);

-- Truyền vào name / tag đã normalize (chữ thường + trim).
-- name: GetPlayerByServerNameAndTag :one
SELECT * FROM lol_players WHERE server = sqlc.arg(server) AND normalized_name = sqlc.arg(normalized_name) AND normalized_tag = sqlc.arg(normalized_tag);

-- name: SearchPlayers :many
SELECT * FROM lol_players
WHERE name ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR tag ILIKE '%' || sqlc.arg(keyword)::text || '%'
   OR search_string LIKE '%' || sqlc.arg(normalized_keyword)::text || '%'
ORDER BY solo_rank_power DESC NULLS LAST
LIMIT sqlc.arg(lim);

-- Match đã kết thúc không đổi nên trùng id thì bỏ qua.
-- name: InsertMatch :exec
INSERT INTO lol_matches (
    id, server, mode, patch, game_start_at, duration_sec, is_remake, estimated_rank,
    banned_champion_ids,
    winning_team,
    team_1_kills, team_1_dragon_kills, team_1_herald_kills, team_1_baron_kills,
    team_2_kills, team_2_dragon_kills, team_2_herald_kills, team_2_baron_kills
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT (id) DO NOTHING;

-- name: InsertMatchParticipant :exec
INSERT INTO lol_match_participants (
    match_id, team, is_win, player_id,
    riot_name, riot_tag, rank_power,
    champion_id, champ_level, position,
    kills, deaths, assists, kda, kill_participation,
    gold_earned, minions_killed, neutral_minions_killed, cs,
    dmg_to_champs, physical_dmg_to_champs, magic_dmg_to_champs, true_dmg_to_champs, dmg_taken, vision_score,
    perf_score,
    spell1_id, spell2_id,
    rune_primary_style, rune_sub_style, key_rune, runes, stat_runes,
    items
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
    $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
)
ON CONFLICT (match_id, player_id) DO NOTHING;

-- Bản batch của InsertMatch: pgx gửi cả list trong 1 round trip.
-- name: InsertMatches :batchexec
INSERT INTO lol_matches (
    id, server, mode, patch, game_start_at, duration_sec, is_remake, estimated_rank,
    banned_champion_ids,
    winning_team,
    team_1_kills, team_1_dragon_kills, team_1_herald_kills, team_1_baron_kills,
    team_2_kills, team_2_dragon_kills, team_2_herald_kills, team_2_baron_kills
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
ON CONFLICT (id) DO NOTHING;

-- Bản batch của InsertMatchParticipant.
-- name: InsertMatchParticipants :batchexec
INSERT INTO lol_match_participants (
    match_id, team, is_win, player_id,
    riot_name, riot_tag, rank_power,
    champion_id, champ_level, position,
    kills, deaths, assists, kda, kill_participation,
    gold_earned, minions_killed, neutral_minions_killed, cs,
    dmg_to_champs, physical_dmg_to_champs, magic_dmg_to_champs, true_dmg_to_champs, dmg_taken, vision_score,
    perf_score,
    spell1_id, spell2_id,
    rune_primary_style, rune_sub_style, key_rune, runes, stat_runes,
    items
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
    $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
)
ON CONFLICT (match_id, player_id) DO NOTHING;

-- Id không có trong DB thì bỏ qua.
-- name: GetMatchesByIds :many
SELECT * FROM lol_matches
WHERE id = ANY(sqlc.arg(ids)::text[])
ORDER BY game_start_at DESC;

-- name: GetMatchParticipantsByMatchIds :many
SELECT * FROM lol_match_participants
WHERE match_id = ANY(sqlc.arg(match_ids)::text[])
ORDER BY match_id, team, array_position(ARRAY['TOP', 'JGL', 'MID', 'ADC', 'SPT'], position), player_id;
