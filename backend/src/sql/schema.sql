CREATE TABLE champions (
    id          INTEGER PRIMARY KEY,               -- ddragon "key" ("266") = championId trong match
    slug        TEXT        NOT NULL UNIQUE,       -- ddragon "id" ("Aatrox"), dùng cho URL ảnh / tên file
    name        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    img_url     TEXT        NOT NULL,              -- URL đầy đủ tới ảnh trên ddragon
    version     TEXT        NOT NULL,              -- patch lúc sync, vd "16.18.1"
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE items (
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
