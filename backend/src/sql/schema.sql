CREATE TABLE IF NOT EXISTS lol_champions (
    id          INTEGER PRIMARY KEY,               -- ddragon "key" ("266") = championId trong match
    slug        TEXT        NOT NULL UNIQUE,       -- ddragon "id" ("Aatrox"), dùng cho URL ảnh / tên file
    name        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    patch       TEXT        NOT NULL,              -- patch lúc sync, vd "16.18" (cắt từ ddragon version "16.18.1")
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lol_items (
    id          INTEGER PRIMARY KEY,               -- item id = item0..item6 trong match
    name        TEXT        NOT NULL,
    plaintext   TEXT        NOT NULL DEFAULT '',
    type        TEXT        NOT NULL,              -- CONSUMABLE | TRINKET | BOOTS | STARTER | BASIC | EPIC | LEGENDARY
    gold_total  INTEGER     NOT NULL,
    from_items  INTEGER[]   NOT NULL DEFAULT '{}', -- item id của các thành phần
    into_items  INTEGER[]   NOT NULL DEFAULT '{}', -- item id nâng cấp lên
    is_summoners_rift BOOLEAN NOT NULL,            -- dùng được ở Summoner's Rift (map 11)
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    patch       TEXT        NOT NULL,              -- patch lúc sync, vd "16.18" (cắt từ ddragon version "16.18.1")
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lol_spells (
    id          INTEGER PRIMARY KEY,               -- ddragon "key" ("4") = spell1_id / spell2_id trong match
    slug        TEXT        NOT NULL UNIQUE,       -- ddragon "id" ("SummonerFlash")
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    patch       TEXT        NOT NULL,              -- patch lúc sync, vd "16.18"
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Gộp cả cây rune lẫn rune: hai thứ dùng chung bộ cột, chỉ khác ở chỗ cây không có cha và không có slot.
CREATE TABLE IF NOT EXISTS lol_runes (
    id          INTEGER PRIMARY KEY,               -- cây: = rune_primary_style / rune_sub_style trong match
                                                   -- rune: = key_rune và các phần tử của runes
    style_id    INTEGER     REFERENCES lol_runes (id) ON DELETE CASCADE,
                                                   -- NULL = dòng này LÀ một cây; khác NULL = rune thuộc cây đó
    slot        INTEGER,                           -- hàng trong cây, 0 = hàng keystone; NULL với cây
    slug        TEXT        NOT NULL UNIQUE,       -- ddragon "key": "Precision" | "Electrocute"
    name        TEXT        NOT NULL,
    short_desc  TEXT        NOT NULL DEFAULT '',   -- luôn rỗng với cây
    img_url     TEXT        NOT NULL,              -- icon rune dùng path không kèm version: /cdn/img/perk-images/...
    patch       TEXT        NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS lol_runes_style_slot_idx ON lol_runes (style_id, slot);

CREATE TABLE IF NOT EXISTS lol_players (
    id                 TEXT        PRIMARY KEY,           -- puuid, định danh duy nhất toàn cầu của Riot
    server             TEXT        NOT NULL,              -- enum Server của app: VN | KR | EUW | NA ... (KHÔNG phải platform id của Riot)
    name               TEXT        NOT NULL,              -- Riot ID gameName, lấy từ account-v1
    tag                TEXT        NOT NULL,              -- Riot ID tagLine
    normalized_name    TEXT        NOT NULL,              -- name chữ thường + trim, giữ dấu; dùng để find chính xác
    normalized_tag     TEXT        NOT NULL,              -- tag chữ thường + trim, giữ dấu
    profile_icon_id    INTEGER,                           -- từ summoner-v4
    level              INTEGER,                           -- summonerLevel của Riot
    search_string      TEXT        NOT NULL DEFAULT '',   -- NormalizeString(normalized_name) + "#" + NormalizeString(normalized_tag), bỏ dấu

    -- rank solo/duo (RANKED_SOLO_5x5); rank = Riot "tier", tier = Riot "rank"
    solo_rank          TEXT,                              -- CHALLENGER | GRANDMASTER | MASTER | DIAMOND ... | UNRANKED; NULL = unknown (chưa fetch league)
    solo_tier          TEXT,                              -- I | II | III | IV; NULL = unknown
    solo_lp            INTEGER     NOT NULL DEFAULT 0,
    solo_rank_power    INTEGER,                           -- bậc rank * 400 + bậc tier * 100 + lp; NULL = unknown, 0 = unranked
    solo_wins          INTEGER     NOT NULL DEFAULT 0,
    solo_losses        INTEGER     NOT NULL DEFAULT 0,

    -- rank flex (RANKED_FLEX_SR)
    flex_rank          TEXT,                              -- NULL = unknown
    flex_tier          TEXT,                              -- NULL = unknown
    flex_lp            INTEGER     NOT NULL DEFAULT 0,
    flex_wins          INTEGER     NOT NULL DEFAULT 0,
    flex_losses        INTEGER     NOT NULL DEFAULT 0,

    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS lol_players_normalized_name_tag_idx ON lol_players (normalized_name, normalized_tag);

CREATE TABLE IF NOT EXISTS lol_matches (
    id              TEXT        PRIMARY KEY,           -- metadata.matchId, vd "VN2_123456789"
    server          TEXT        NOT NULL,              -- enum Server của app, map từ info.platformId
    mode            TEXT        NOT NULL,              -- enum GameMode của app, gộp từ queueId + mapId: SOLO | FLEX | ARAM | NORMAL
    patch           TEXT        NOT NULL,              -- "16.18", cắt từ info.gameVersion ("16.18.712.3456") trong app
    game_start_at   TIMESTAMPTZ NOT NULL,
    duration_sec    INTEGER     NOT NULL,
    is_remake       BOOLEAN     NOT NULL,              -- gameEndedInEarlySurrender
    estimated_rank  TEXT        NOT NULL,              -- CHALLENGER | GRANDMASTER | MASTER | DIAMOND ...; UNRANKED = không participant nào có rank đã biết

    banned_champion_ids  INTEGER[] NOT NULL,           -- chỉ các tướng bị ban, bỏ lượt không ban (-1)

    -- team: 1 = blue (Riot 100) | 2 = red (Riot 200)
    winning_team         SMALLINT  NOT NULL,
    team_1_kills         SMALLINT  NOT NULL,
    team_1_dragon_kills  SMALLINT  NOT NULL,
    team_1_herald_kills  SMALLINT  NOT NULL,
    team_1_baron_kills   SMALLINT  NOT NULL,
    team_2_kills         SMALLINT  NOT NULL,
    team_2_dragon_kills  SMALLINT  NOT NULL,
    team_2_herald_kills  SMALLINT  NOT NULL,
    team_2_baron_kills   SMALLINT  NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS lol_matches_filter_idx ON lol_matches (patch, mode, game_start_at DESC);

CREATE TABLE IF NOT EXISTS lol_match_participants (
    match_id               TEXT      NOT NULL REFERENCES lol_matches (id) ON DELETE CASCADE,
    team                   SMALLINT  NOT NULL,         -- 1 = blue (Riot 100) | 2 = red (Riot 200)
    is_win                 BOOLEAN   NOT NULL,
    player_id              TEXT      NOT NULL,         -- puuid; KHÔNG FK tới lol_players vì không phải ai cũng được crawl

    name                   TEXT      NOT NULL,         -- snapshot Riot ID lúc chơi
    tag                    TEXT      NOT NULL,
    rank_power             INTEGER,                    -- snapshot solo_rank_power lúc insert; NULL = unknown

    champion_id            INTEGER   NOT NULL,
    champion_slug          TEXT      NOT NULL,         -- snapshot championName của Riot ("Aatrox", "MonkeyKing") = lol_champions.slug
    champ_level            SMALLINT  NOT NULL,
    position               TEXT      NOT NULL,         -- TOP | JGL | MID | ADC | SPT (map từ teamPosition) | UNK nếu rỗng (ARAM...)

    kills                  SMALLINT  NOT NULL,
    deaths                 SMALLINT  NOT NULL,
    assists                SMALLINT  NOT NULL,
    kda                    REAL      NOT NULL,         -- (kills + assists) / max(deaths, 1), tính trong app
    kill_participation     REAL      NOT NULL,         -- (kills + assists) / team kills, tính trong app; 0 nếu team 0 kill
    double_kills           SMALLINT  NOT NULL,
    triple_kills           SMALLINT  NOT NULL,
    quadra_kills           SMALLINT  NOT NULL,
    penta_kills            SMALLINT  NOT NULL,
    solo_kills             SMALLINT  NOT NULL,         -- challenges.soloKills: hạ gục không có đồng đội tham gia

    gold                   INTEGER   NOT NULL,
    gold_per_min           REAL      NOT NULL,         -- gold / phút, tính trong app
    minions_killed         INTEGER   NOT NULL,
    neutral_minions_killed INTEGER   NOT NULL,
    cs                     INTEGER   NOT NULL,         -- minions_killed + neutral_minions_killed, tính trong app
    cs_per_min             REAL      NOT NULL,         -- cs / phút, tính trong app

    dmg_dealt              INTEGER   NOT NULL,         -- totalDamageDealtToChampions
    dmg_per_min            REAL      NOT NULL,         -- dmg_dealt / phút, tính trong app
    physical_dmg_dealt     INTEGER   NOT NULL,         -- physicalDamageDealtToChampions
    magic_dmg_dealt        INTEGER   NOT NULL,         -- magicDamageDealtToChampions
    true_dmg_dealt         INTEGER   NOT NULL,         -- trueDamageDealtToChampions
    dmg_to_turrets         INTEGER   NOT NULL,         -- damageDealtToTurrets
    dmg_taken              INTEGER   NOT NULL,         -- totalDamageTaken
    heal                   INTEGER   NOT NULL,         -- totalHeal, gồm cả tự hồi (lifesteal, hồi máu, bình máu)
    heal_others            INTEGER   NOT NULL,         -- totalHealsOnTeammates, chỉ phần hồi cho đồng đội
    shield_others          INTEGER   NOT NULL,         -- totalDamageShieldedOnTeammates
    vision_score           INTEGER   NOT NULL,
    wards_placed           INTEGER   NOT NULL,
    wards_killed           INTEGER   NOT NULL,

    perf_score             INTEGER   NOT NULL,         -- điểm hiệu suất, công thức riêng tính trong app

    -- spell lưu đã sort (spell1 < spell2) để Flash+Ignite == Ignite+Flash khi GROUP BY
    spell1_id              SMALLINT  NOT NULL,
    spell2_id              SMALLINT  NOT NULL,

    rune_primary_style     INTEGER   NOT NULL,
    rune_sub_style         INTEGER   NOT NULL,
    key_rune               INTEGER   NOT NULL,         -- rune đầu tiên của primary style
    runes                  INTEGER[] NOT NULL,         -- 4 primary + 2 sub, đúng thứ tự
    stat_runes             INTEGER[] NOT NULL,         -- [offense, flex, defense]

    items                  INTEGER[] NOT NULL,         -- đủ 7 phần tử item0..item6, giữ 0 cho slot rỗng; phần tử thứ 7 = trinket

    PRIMARY KEY (match_id, player_id)
);

CREATE INDEX IF NOT EXISTS lol_match_participants_player_idx   ON lol_match_participants (player_id);
CREATE INDEX IF NOT EXISTS lol_match_participants_champion_idx ON lol_match_participants (champion_id, position);











-- ============================================================
-- ANALYTICS — bảng tổng hợp, chỉ job ghi. Xóa sạch rồi chạy lại
-- job phải ra đúng kết quả cũ (idempotent, không có watermark).
-- ============================================================

-- Một "meta" = một lát cắt tổng hợp theo (patch, server, rank_bucket).
CREATE TABLE IF NOT EXISTS lol_metas (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(), -- job khỏi tự sinh; upsert theo UNIQUE bên dưới nên id giữ nguyên qua mỗi lần refresh
    patch         TEXT        NOT NULL,
    server        TEXT        NOT NULL,   -- GLOBAL + các server enum trong app: GLOBAL | VN | KR ...
    rank_bucket   TEXT        NOT NULL,   -- MASTER_PLUS = estimated_rank thuộc MASTER | GRANDMASTER | CHALLENGER; có thể tách riêng 3 bậc khi đủ mẫu
    total_matches INTEGER     NOT NULL,   -- đã lọc remake / trận ngắn / non-ranked; MẪU SỐ chung cho pick_rate & ban_rate
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (patch, server, rank_bucket)   -- job upsert theo khóa này
);

CREATE TABLE IF NOT EXISTS lol_champion_stats (
    meta_id     UUID    NOT NULL REFERENCES lol_metas (id) ON DELETE CASCADE,
    position    TEXT    NOT NULL,         -- TOP | JGL | MID | ADC | SPT
    champion_id INTEGER NOT NULL,         -- không FK tới lol_champions: tướng mới có thể vào match trước khi ddragon sync; tên/ảnh join lúc đọc

    -- Số đếm, không phải tỉ lệ: cần để gộp nhiều dòng lại và để tính Wilson score.
    -- power (điểm xếp hạng) & tier (S/A/B...) KHÔNG lưu ở đây: đều suy ra được từ các cột
    -- dưới đây trên chính 1 dòng, tính lúc đọc để đổi công thức/ngưỡng khỏi chạy lại job.
    games     INTEGER NOT NULL,
    wins      INTEGER NOT NULL,
    win_rate  DOUBLE PRECISION NOT NULL,          -- wins / games; lưu sẵn để sort/lọc, đọc khỏi tính
    pick_rate DOUBLE PRECISION NOT NULL,          -- games / meta.total_matches

    -- Trung bình có trọng số. Gộp dòng: SUM(avg_x * games) / SUM(games). KHÔNG PHẢI AVG(avg_x).
    avg_kills            DOUBLE PRECISION NOT NULL,
    avg_deaths           DOUBLE PRECISION NOT NULL,
    avg_assists          DOUBLE PRECISION NOT NULL,
    avg_kda              DOUBLE PRECISION NOT NULL,
    avg_kp               DOUBLE PRECISION NOT NULL,
    avg_cs_per_min       DOUBLE PRECISION NOT NULL,
    avg_gold_per_min     DOUBLE PRECISION NOT NULL,
    avg_dmg_per_min      DOUBLE PRECISION NOT NULL,
    avg_physical_dmg     DOUBLE PRECISION NOT NULL,
    avg_magic_dmg        DOUBLE PRECISION NOT NULL,
    avg_true_dmg         DOUBLE PRECISION NOT NULL,
    avg_penta            DOUBLE PRECISION NOT NULL,
    avg_solo_kills       DOUBLE PRECISION NOT NULL,
    avg_perf_score       DOUBLE PRECISION NOT NULL,

    -- Build tốt nhất, mỗi phần tử LƯU CẢ COUNT (games/wins) để recompute rate & pure-SQL được.
    -- best_spell_combos:    [{"s1":Int,"s2":Int,"games":Int,"wins":Int}]  -- spell đã sort s1<s2
    -- best_runes:           [{"main_style":Int,"sub_style":Int,"key_rune":Int,"runes":[Int],"stat_runes":[Int],"games":Int,"wins":Int}]
    -- best_legendary_items: [{"item_id":Int,"games":Int,"wins":Int}]      -- lol_items.type = LEGENDARY
    -- best_boot_items:      [{"item_id":Int,"games":Int,"wins":Int}]      -- lol_items.type = BOOTS
    -- matchups:             [{"opponent_champion_id":Int,"games":Int,"wins":Int}]  -- đối đầu cùng lane, khác phe
    best_spell_combos    JSONB NOT NULL DEFAULT '[]',
    best_runes           JSONB NOT NULL DEFAULT '[]',
    best_legendary_items JSONB NOT NULL DEFAULT '[]',
    best_boot_items      JSONB NOT NULL DEFAULT '[]',
    matchups             JSONB NOT NULL DEFAULT '[]',

    PRIMARY KEY (meta_id, position, champion_id)
);

-- Tier list: lọc meta + position, sắp theo cỡ mẫu.
CREATE INDEX IF NOT EXISTS lol_champion_stats_read_idx
    ON lol_champion_stats (meta_id, position, games DESC);

-- Ban ở cấp TRẬN, không gắn lane => tách riêng theo (meta, champion) thay vì lặp trên mọi dòng
-- position như bản cũ (hết footgun MAX-vs-SUM). Presence = pick (champion_stats) + ban (bảng này).
CREATE TABLE IF NOT EXISTS lol_champion_bans (
    meta_id     UUID    NOT NULL REFERENCES lol_metas (id) ON DELETE CASCADE,
    champion_id INTEGER NOT NULL,
    bans        INTEGER NOT NULL,
    ban_rate    DOUBLE PRECISION NOT NULL,        -- bans / meta.total_matches

    PRIMARY KEY (meta_id, champion_id)
);