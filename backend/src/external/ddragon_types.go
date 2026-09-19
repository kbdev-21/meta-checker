package external

type ddResponse[T any] struct {
	Type    string       `json:"type"`
	Version string       `json:"version"`
	Data    map[string]T `json:"data"`
}

type DDImage struct {
	Full   string `json:"full"`
	Sprite string `json:"sprite"`
	Group  string `json:"group"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
}

// ---------- champion.json ----------

type DDChampion struct {
	Id      string             `json:"id"`  // "Aatrox"
	Key     string             `json:"key"` // "266" = championId
	Name    string             `json:"name"`
	Title   string             `json:"title"`
	Blurb   string             `json:"blurb"`
	Tags    []string           `json:"tags"`
	Partype string             `json:"partype"`
	Info    DDChampionInfo     `json:"info"`
	Image   DDImage            `json:"image"`
	Stats   map[string]float64 `json:"stats"`
}

type DDChampionInfo struct {
	Attack     int `json:"attack"`
	Defense    int `json:"defense"`
	Magic      int `json:"magic"`
	Difficulty int `json:"difficulty"`
}

// ---------- item.json ----------

type DDItem struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Plaintext   string             `json:"plaintext"`
	From        []string           `json:"from"`
	Into        []string           `json:"into"`
	Depth       int                `json:"depth"`
	Tags        []string           `json:"tags"`
	Maps        map[string]bool    `json:"maps"` // mapId -> có dùng được không (11 = SR)
	Gold        DDItemGold         `json:"gold"`
	Stats       map[string]float64 `json:"stats"`
	Image       DDImage            `json:"image"`
}

type DDItemGold struct {
	Base        int  `json:"base"`
	Total       int  `json:"total"`
	Sell        int  `json:"sell"`
	Purchasable bool `json:"purchasable"`
}

// ---------- summoner.json ----------

type DDSummonerSpell struct {
	Id            string    `json:"id"`  // "SummonerFlash"
	Key           string    `json:"key"` // "4" = summoner1Id/2Id
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Cooldown      []float64 `json:"cooldown"`
	SummonerLevel int       `json:"summonerLevel"`
	Modes         []string  `json:"modes"`
	Image         DDImage   `json:"image"`
}

// ---------- runesReforged.json ----------

type DDRuneTree struct {
	Id    int          `json:"id"` // = PerkStyleDto.Style
	Key   string       `json:"key"`
	Icon  string       `json:"icon"` // path dưới /cdn/img/
	Name  string       `json:"name"`
	Slots []DDRuneSlot `json:"slots"`
}

type DDRuneSlot struct {
	Runes []DDRune `json:"runes"`
}

type DDRune struct {
	Id        int    `json:"id"` // = PerkStyleSelectionDto.Perk
	Key       string `json:"key"`
	Icon      string `json:"icon"`
	Name      string `json:"name"`
	ShortDesc string `json:"shortDesc"`
	LongDesc  string `json:"longDesc"`
}
