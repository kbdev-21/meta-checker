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
    is_sr       BOOLEAN     NOT NULL,              -- dùng được ở Summoner's Rift (map 11)
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    patch       TEXT        NOT NULL,              -- patch lúc sync, vd "16.18" (cắt từ ddragon version "16.18.1")
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lol_players (
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

CREATE INDEX IF NOT EXISTS lol_players_matches_synced_at_idx ON lol_players (matches_synced_at NULLS FIRST);
CREATE INDEX IF NOT EXISTS lol_players_normalized_name_tag_idx ON lol_players (normalized_name, normalized_tag);

CREATE TABLE IF NOT EXISTS lol_matches (
    id              TEXT        PRIMARY KEY,           -- metadata.matchId, vd "VN2_123456789"
    server          TEXT        NOT NULL,              -- info.platformId: VN2 | KR ...
    mode            TEXT        NOT NULL,              -- gộp queueId + mapId + gameMode, enum trong app: RANKED_SOLO | RANKED_FLEX | NORMAL_DRAFT | ARAM ...
    patch           TEXT        NOT NULL,              -- "16.18", cắt từ info.gameVersion ("16.18.712.3456") trong app
    game_start_at   TIMESTAMPTZ NOT NULL,
    duration_sec    INTEGER     NOT NULL,
    is_remake       BOOLEAN     NOT NULL,              -- gameEndedInEarlySurrender
    estimated_rank  TEXT,                              -- CHALLENGER | GRANDMASTER | MASTER | DIAMOND ...; NULL = unknown

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

    riot_name              TEXT      NOT NULL,         -- snapshot Riot ID lúc chơi
    riot_tag               TEXT      NOT NULL,
    rank_power             INTEGER,                    -- snapshot solo_rank_power lúc insert; NULL = unknown

    champion_id            INTEGER   NOT NULL,
    champ_level            SMALLINT  NOT NULL,
    position               TEXT      NOT NULL,         -- TOP | JGL | MID | ADC | SPT (map từ teamPosition) | UNK nếu rỗng (ARAM...)

    kills                  SMALLINT  NOT NULL,
    deaths                 SMALLINT  NOT NULL,
    assists                SMALLINT  NOT NULL,
    kda                    REAL      NOT NULL,         -- (kills + assists) / max(deaths, 1), tính trong app
    kill_participation     REAL      NOT NULL,         -- (kills + assists) / team kills, tính trong app; 0 nếu team 0 kill

    gold_earned            INTEGER   NOT NULL,
    minions_killed         INTEGER   NOT NULL,
    neutral_minions_killed INTEGER   NOT NULL,
    cs                     INTEGER   NOT NULL,         -- minions_killed + neutral_minions_killed, tính trong app

    dmg_to_champs          INTEGER   NOT NULL,         -- totalDamageDealtToChampions
    physical_dmg_to_champs INTEGER   NOT NULL,         -- physicalDamageDealtToChampions
    magic_dmg_to_champs    INTEGER   NOT NULL,         -- magicDamageDealtToChampions
    true_dmg_to_champs     INTEGER   NOT NULL,         -- trueDamageDealtToChampions
    dmg_taken              INTEGER   NOT NULL,         -- totalDamageTaken
    vision_score           INTEGER   NOT NULL,

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
