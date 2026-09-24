-- name: UpsertChampion :exec
INSERT INTO lol_champions (id, slug, name, title, img_url, skills, patch)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    slug       = EXCLUDED.slug,
    name       = EXCLUDED.name,
    title      = EXCLUDED.title,
    img_url    = EXCLUDED.img_url,
    skills     = EXCLUDED.skills,
    patch      = EXCLUDED.patch,
    updated_at = now();

-- name: UpsertItem :exec
INSERT INTO lol_items (id, name, plaintext, type, gold_total, from_items, into_items, is_summoners_rift, img_url, patch)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name       = EXCLUDED.name,
    plaintext  = EXCLUDED.plaintext,
    type       = EXCLUDED.type,
    gold_total = EXCLUDED.gold_total,
    from_items = EXCLUDED.from_items,
    into_items = EXCLUDED.into_items,
    is_summoners_rift = EXCLUDED.is_summoners_rift,
    img_url    = EXCLUDED.img_url,
    patch      = EXCLUDED.patch,
    updated_at = now();

-- name: ListChampions :many
SELECT * FROM lol_champions ORDER BY name;

-- name: ListItems :many
SELECT * FROM lol_items ORDER BY id;

-- name: UpsertSpell :exec
INSERT INTO lol_spells (id, slug, name, description, img_url, patch)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    slug        = EXCLUDED.slug,
    name        = EXCLUDED.name,
    description = EXCLUDED.description,
    img_url     = EXCLUDED.img_url,
    patch       = EXCLUDED.patch,
    updated_at  = now();

-- Cây (style_id NULL) phải upsert trước rune của nó vì style_id tham chiếu ngược về chính bảng này.
-- name: UpsertRune :exec
INSERT INTO lol_runes (id, style_id, slot, sort_order, slug, name, short_desc, img_url, patch)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
    style_id   = EXCLUDED.style_id,
    slot       = EXCLUDED.slot,
    sort_order = EXCLUDED.sort_order,
    slug       = EXCLUDED.slug,
    name       = EXCLUDED.name,
    short_desc = EXCLUDED.short_desc,
    img_url    = EXCLUDED.img_url,
    patch      = EXCLUDED.patch,
    updated_at = now();

-- name: ListSpells :many
SELECT * FROM lol_spells ORDER BY id;

-- Cây trước, rồi tới rune của từng cây theo đúng thứ tự hàng; trong 1 hàng theo thứ tự trong game.
-- name: ListRunes :many
SELECT * FROM lol_runes ORDER BY style_id NULLS FIRST, slot NULLS FIRST, sort_order;

-- Rank solo/flex truyền vào NULL (unknown) thì giữ nguyên cả nhóm cột rank cũ.
-- name: UpsertPlayer :exec
INSERT INTO lol_players (
    id, server, name, tag, normalized_name, normalized_tag, profile_icon_id, level, search_string,
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
    level           = COALESCE(EXCLUDED.level, lol_players.level),
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

-- Batch: pgx gửi cả list trong 1 round trip. Match đã kết thúc không đổi nên trùng id thì bỏ qua.
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

-- Batch, trùng (match_id, player_id) thì bỏ qua.
-- name: InsertMatchParticipants :batchexec
INSERT INTO lol_match_participants (
    match_id, team, is_win, player_id, participant_id,
    name, tag, rank_power,
    champion_id, champion_slug, champ_level, position,
    kills, deaths, assists, kda, kill_participation,
    double_kills, triple_kills, quadra_kills, penta_kills, solo_kills,
    gold, gold_per_min, minions_killed, neutral_minions_killed, cs, cs_per_min,
    dmg_dealt, dmg_per_min, physical_dmg_dealt, magic_dmg_dealt, true_dmg_dealt, dmg_to_turrets,
    dmg_taken, dmg_taken_per_min, crowd_control, cc_per_min,
    heal, heal_others, shield_others,
    vision_score, wards_placed, wards_killed,
    perf_score,
    spell1_id, spell2_id,
    rune_primary_style, rune_sub_style, key_rune, runes, stat_runes,
    items,
    starter_sets, skills_leveled, first_legend_item, legend_items_purchased
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39,
    $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57
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


-- ============================================================
-- ANALYTICS
-- Lọc match hợp lệ của 1 lát cắt luôn cùng bộ điều kiện: patch + mode ranked +
-- không remake + đủ dài + đúng rank bucket + (GLOBAL hoặc đúng server).
-- ============================================================

-- name: CountSliceMatches :one
SELECT count(*) FROM lol_matches m
WHERE m.patch = sqlc.arg(patch)::text
  AND m.mode = 'SOLO'
  AND NOT m.is_remake
  AND m.duration_sec >= sqlc.arg(min_duration)::int
  AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
  AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text);

-- Upsert theo khóa tự nhiên: id giữ nguyên qua mỗi lần refresh.
-- name: UpsertMeta :one
INSERT INTO lol_metas (patch, server, rank_bucket, total_matches)
VALUES (sqlc.arg(patch)::text, sqlc.arg(server)::text, sqlc.arg(rank_bucket)::text, sqlc.arg(total_matches)::int)
ON CONFLICT (patch, server, rank_bucket) DO UPDATE SET
    total_matches = EXCLUDED.total_matches,
    updated_at    = now()
RETURNING id;

-- name: GetMeta :one
SELECT * FROM lol_metas
WHERE patch = sqlc.arg(patch)::text AND server = sqlc.arg(server)::text AND rank_bucket = sqlc.arg(rank_bucket)::text;

-- name: DeleteChampionStatsByMeta :exec
DELETE FROM lol_champion_stats WHERE meta_id = sqlc.arg(meta_id)::uuid;

-- name: DeleteChampionBansByMeta :exec
DELETE FROM lol_champion_bans WHERE meta_id = sqlc.arg(meta_id)::uuid;

-- Scalar aggregate. Cột JSONB để DEFAULT '[]', các query Refresh...* bên dưới điền sau.
-- avg_x lưu trung bình có trọng số theo participant; gộp nhiều dòng sau này dùng SUM(avg_x*games)/SUM(games).
-- name: RefreshChampionStatsScalars :exec
INSERT INTO lol_champion_stats (
    meta_id, position, champion_id, champion_slug,
    games, wins, win_rate, pick_rate,
    avg_kills, avg_deaths, avg_assists, avg_kda, avg_kp,
    avg_cs_per_min, avg_gold_per_min, avg_dmg_per_min, avg_dmg_taken_per_min, avg_cc_per_min,
    avg_physical_dmg, avg_magic_dmg, avg_true_dmg,
    avg_penta, avg_solo_kills, avg_perf_score
)
SELECT
    sqlc.arg(meta_id)::uuid, mp.position, mp.champion_id, max(mp.champion_slug),
    count(*),
    count(*) FILTER (WHERE mp.is_win),
    count(*) FILTER (WHERE mp.is_win)::float8 / count(*),
    count(*)::float8 / nullif(sqlc.arg(total_matches)::int, 0),
    avg(mp.kills), avg(mp.deaths), avg(mp.assists), avg(mp.kda), avg(mp.kill_participation),
    avg(mp.cs_per_min), avg(mp.gold_per_min), avg(mp.dmg_per_min), avg(mp.dmg_taken_per_min), avg(mp.cc_per_min),
    avg(mp.physical_dmg_dealt), avg(mp.magic_dmg_dealt), avg(mp.true_dmg_dealt),
    avg(mp.penta_kills), avg(mp.solo_kills), avg(mp.perf_score)
FROM (
    SELECT p.*
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
) mp
GROUP BY mp.position, mp.champion_id;

-- name: RefreshChampionStatsSpellCombos :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win, p.spell1_id, p.spell2_id
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
),
agg AS (
    SELECT position, champion_id, spell1_id, spell2_id,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, spell1_id, spell2_id
)
UPDATE lol_champion_stats cs
SET best_spell_combos = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'spell1Id', spell1_id, 'spell2Id', spell2_id, 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- name: RefreshChampionStatsRunes :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win,
           p.rune_primary_style, p.rune_sub_style, p.key_rune, p.runes, p.stat_runes
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
),
agg AS (
    SELECT position, champion_id, rune_primary_style, rune_sub_style, key_rune, runes, stat_runes,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, rune_primary_style, rune_sub_style, key_rune, runes, stat_runes
)
UPDATE lol_champion_stats cs
SET best_runes = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'runePrimaryStyle', rune_primary_style, 'runeSubStyle', rune_sub_style,
               'keyRune', key_rune, 'runes', to_jsonb(runes), 'statRunes', to_jsonb(stat_runes),
               'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Item legendary: unnest items rồi lọc theo lol_items.type. Item biến đổi (Muramana...) gộp về
-- item gốc (Manamune...) theo cặp transformed_ids[i] => base_ids[i] app truyền vào.
-- name: RefreshChampionStatsLegendaryItems :exec
WITH item_rows AS (
    SELECT p.position, p.champion_id, p.is_win, it.id AS item_id
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    CROSS JOIN LATERAL unnest(p.items) AS iid
    -- iid không có trong transformed_ids => array_position NULL => phần tử NULL => giữ iid.
    JOIN lol_items it
        ON it.id = COALESCE((sqlc.arg(base_ids)::int[])[array_position(sqlc.arg(transformed_ids)::int[], iid)], iid)
       AND it.type = 'LEGENDARY'
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
),
agg AS (
    SELECT position, champion_id, item_id,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM item_rows
    GROUP BY position, champion_id, item_id
)
UPDATE lol_champion_stats cs
SET best_legendary_items = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'itemId', item_id, 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- name: RefreshChampionStatsBootItems :exec
WITH item_rows AS (
    SELECT p.position, p.champion_id, p.is_win, it.id AS item_id
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    CROSS JOIN LATERAL unnest(p.items) AS iid
    JOIN lol_items it ON it.id = iid AND it.type = 'BOOTS'
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
),
agg AS (
    SELECT position, champion_id, item_id,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM item_rows
    GROUP BY position, champion_id, item_id
)
UPDATE lol_champion_stats cs
SET best_boot_items = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'itemId', item_id, 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Matchup: self-join participant cùng match, cùng position, khác phe.
-- name: RefreshChampionStatsMatchups :exec
WITH mp AS (
    SELECT p.match_id, p.team, p.position, p.champion_id, p.is_win
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
),
pairs AS (
    SELECT me.position, me.champion_id, opp.champion_id AS opponent_champion_id, me.is_win
    FROM mp me
    JOIN mp opp ON opp.match_id = me.match_id AND opp.position = me.position AND opp.team <> me.team
),
agg AS (
    SELECT position, champion_id, opponent_champion_id,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM pairs
    GROUP BY position, champion_id, opponent_champion_id
)
UPDATE lol_champion_stats cs
SET best_match_ups = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'opponentChampionId', opponent_champion_id, 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Các cột từ timeline: chỉ tính participant có dữ liệu tương ứng (không có timeline => mảng rỗng / 0).
-- Starter set đã sort lúc lưu nên GROUP BY thẳng mảng.
-- name: RefreshChampionStatsStarterSets :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win, p.starter_sets AS item_ids
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
      AND cardinality(p.starter_sets) > 0
),
agg AS (
    SELECT position, champion_id, item_ids,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, item_ids
)
UPDATE lol_champion_stats cs
SET best_starter_sets = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'itemIds', to_jsonb(item_ids), 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Gộp theo 13 lần lên skill đầu: vừa đủ max 2 skill thường, phần sau phân mảnh theo độ dài trận.
-- name: RefreshChampionStatsSkillsLeveled :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win, p.skills_leveled[1:13] AS skills
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
      AND cardinality(p.skills_leveled) >= 13
),
agg AS (
    SELECT position, champion_id, skills,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, skills
)
UPDATE lol_champion_stats cs
SET best_skills_leveled = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'skills', to_jsonb(skills), 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Đồ legendary hoàn thành đầu tiên.
-- name: RefreshChampionStatsFirstLegendItems :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win, p.first_legend_item AS item_id
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
      AND p.first_legend_item <> 0
),
agg AS (
    SELECT position, champion_id, item_id,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, item_id
)
UPDATE lol_champion_stats cs
SET best_first_legend_items = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'itemId', item_id, 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- 3 đồ legendary đầu, GIỮ thứ tự mua (A>B>C khác B>A>C).
-- name: RefreshChampionStatsFirstThreeItems :exec
WITH mp AS (
    SELECT p.position, p.champion_id, p.is_win, p.legend_items_purchased[1:3] AS item_ids
    FROM lol_match_participants p
    JOIN lol_matches m ON m.id = p.match_id
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
      AND p.position <> 'UNK'
      AND cardinality(p.legend_items_purchased) >= 3
),
agg AS (
    SELECT position, champion_id, item_ids,
           count(*) AS games,
           count(*) FILTER (WHERE is_win) AS wins
    FROM mp
    GROUP BY position, champion_id, item_ids
)
UPDATE lol_champion_stats cs
SET best_first_three_items = sub.arr
FROM (
    SELECT position, champion_id,
           jsonb_agg(jsonb_build_object(
               'itemIds', to_jsonb(item_ids), 'games', games, 'wins', wins
           ) ORDER BY games DESC) AS arr
    FROM agg
    GROUP BY position, champion_id
) sub
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.position = sub.position
  AND cs.champion_id = sub.champion_id;

-- Ban ở cấp trận: unnest banned_champion_ids (đã bỏ -1 lúc lưu).
-- name: RefreshChampionBans :exec
INSERT INTO lol_champion_bans (meta_id, champion_id, bans, ban_rate)
SELECT sqlc.arg(meta_id)::uuid, ban.champion_id, count(*),
       count(*)::float8 / nullif(sqlc.arg(total_matches)::int, 0)
FROM (
    SELECT unnest(m.banned_champion_ids) AS champion_id
    FROM lol_matches m
    WHERE m.patch = sqlc.arg(patch)::text
      AND m.mode = 'SOLO'
      AND NOT m.is_remake
      AND m.duration_sec >= sqlc.arg(min_duration)::int
      AND m.estimated_rank = ANY(sqlc.arg(ranks)::text[])
      AND (sqlc.arg(server)::text = 'GLOBAL' OR m.server = sqlc.arg(server)::text)
) ban
WHERE ban.champion_id >= 0
GROUP BY ban.champion_id;

-- Tier list: cột scalar + best_match_ups, KHÔNG lấy các cột JSONB build còn lại (chúng chiếm
-- ~85% payload của cả meta mà tier list không dùng tới); muốn build thì gọi
-- GetChampionStatsByMetaAndChampId cho từng tướng.
-- name: GetChampionStatsByMeta :many
SELECT cs.meta_id, cs.position, cs.champion_id, cs.champion_slug,
       cs.games, cs.wins, cs.win_rate, cs.pick_rate,
       cs.avg_kills, cs.avg_deaths, cs.avg_assists, cs.avg_kda, cs.avg_kp,
       cs.avg_cs_per_min, cs.avg_gold_per_min, cs.avg_dmg_per_min, cs.avg_dmg_taken_per_min, cs.avg_cc_per_min,
       cs.avg_physical_dmg, cs.avg_magic_dmg, cs.avg_true_dmg,
       cs.avg_penta, cs.avg_solo_kills, cs.avg_perf_score,
       cs.best_match_ups,
       COALESCE(b.bans, 0)::int AS bans, COALESCE(b.ban_rate, 0)::float8 AS ban_rate
FROM lol_champion_stats cs
LEFT JOIN lol_champion_bans b ON b.meta_id = cs.meta_id AND b.champion_id = cs.champion_id
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
ORDER BY array_position(ARRAY['TOP', 'JGL', 'MID', 'ADC', 'SPT'], cs.position), cs.games DESC;

-- Như trên nhưng chỉ 1 tướng (mọi position của nó), và có kèm build.
-- name: GetChampionStatsByMetaAndChampId :many
SELECT cs.*, COALESCE(b.bans, 0)::int AS bans, COALESCE(b.ban_rate, 0)::float8 AS ban_rate
FROM lol_champion_stats cs
LEFT JOIN lol_champion_bans b ON b.meta_id = cs.meta_id AND b.champion_id = cs.champion_id
WHERE cs.meta_id = sqlc.arg(meta_id)::uuid
  AND cs.champion_id = sqlc.arg(champion_id)::int
ORDER BY array_position(ARRAY['TOP', 'JGL', 'MID', 'ADC', 'SPT'], cs.position), cs.games DESC;
