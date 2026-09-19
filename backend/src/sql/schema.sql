CREATE TABLE IF NOT EXISTS champions (
    id          INTEGER PRIMARY KEY,               -- ddragon "key" ("266") = championId trong match
    slug        TEXT        NOT NULL UNIQUE,       -- ddragon "id" ("Aatrox"), dùng cho URL ảnh / tên file
    name        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    version     TEXT        NOT NULL,              -- patch lúc sync, vd "16.18.1"
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS items (
    id          INTEGER PRIMARY KEY,               -- item id = item0..item6 trong match
    name        TEXT        NOT NULL,
    plaintext   TEXT        NOT NULL DEFAULT '',
    type        TEXT        NOT NULL,              -- CONSUMABLE | TRINKET | BOOTS | STARTER | BASIC | EPIC | LEGENDARY
    gold_total  INTEGER     NOT NULL,
    from_items  INTEGER[]   NOT NULL DEFAULT '{}', -- item id của các thành phần
    into_items  INTEGER[]   NOT NULL DEFAULT '{}', -- item id nâng cấp lên
    is_sr       BOOLEAN     NOT NULL,              -- dùng được ở Summoner's Rift (map 11)
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    version     TEXT        NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS players (
    id                 TEXT        PRIMARY KEY,           -- puuid, định danh duy nhất toàn cầu của Riot
    server             TEXT        NOT NULL,              -- VN2 | KR | EUW1 | NA1 ... (dùng cho league-v4 / summoner-v4)
    name               TEXT        NOT NULL,              -- Riot ID gameName, lấy từ account-v1
    tag                TEXT        NOT NULL,              -- Riot ID tagLine
    normalized_name    TEXT        NOT NULL,              -- name chữ thường + trim, giữ dấu; dùng để find chính xác
    normalized_tag     TEXT        NOT NULL,              -- tag chữ thường + trim, giữ dấu
    profile_icon_id    INTEGER,                           -- từ summoner-v4
    summoner_level     INTEGER,
    search_string      TEXT        NOT NULL DEFAULT '',   -- NormalizeString(normalized_name) + "#" + NormalizeString(normalized_tag), bỏ dấu

    -- rank solo/duo (RANKED_SOLO_5x5); rank = Riot "tier", tier = Riot "rank"
    solo_rank          TEXT,                              -- CHALLENGER | GRANDMASTER | MASTER | DIAMOND ... | UNRANKED; NULL = unknown (chưa fetch league)
    solo_tier          TEXT,                              -- I | II | III | IV; NULL = unknown
    solo_lp            INTEGER     NOT NULL DEFAULT 0,
    solo_rank_power    INTEGER,                           -- công thức bổ sung sau
    solo_wins          INTEGER     NOT NULL DEFAULT 0,
    solo_losses        INTEGER     NOT NULL DEFAULT 0,

    -- rank flex (RANKED_FLEX_SR)
    flex_rank          TEXT,                              -- NULL = unknown
    flex_tier          TEXT,                              -- NULL = unknown
    flex_lp            INTEGER     NOT NULL DEFAULT 0,
    flex_wins          INTEGER     NOT NULL DEFAULT 0,
    flex_losses        INTEGER     NOT NULL DEFAULT 0,

    -- phục vụ crawl match
    last_match_at      TIMESTAMPTZ,                       -- thời điểm kết thúc của match mới nhất đã lấy, dùng làm startTime cho lần sau
    matches_synced_at  TIMESTAMPTZ,                       -- lần cuối gọi match-v5 cho player này

    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS players_matches_synced_at_idx ON players (matches_synced_at NULLS FIRST);
CREATE INDEX IF NOT EXISTS players_normalized_name_tag_idx ON players (normalized_name, normalized_tag);
