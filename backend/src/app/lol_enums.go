package app

import (
	"slices"
	"strings"

	"backend/src/external"
	"backend/src/shared"
)

// Enum dùng chung của app và các hàm phục vụ chúng.
// Mọi giá trị ở đây là của app, không phải của Riot; việc map sang giá trị Riot cũng nằm trong file này.

// ---------- server ----------

// Server của app. KHÔNG phải platform id của Riot ("VN" chứ không phải "VN2").
type Server string

const (
	ServerBR   Server = "BR"
	ServerEUNE Server = "EUNE"
	ServerEUW  Server = "EUW"
	ServerJP   Server = "JP"
	ServerKR   Server = "KR"
	ServerLAN  Server = "LAN"
	ServerLAS  Server = "LAS"
	ServerME   Server = "ME"
	ServerNA   Server = "NA"
	ServerOCE  Server = "OCE"
	ServerRU   Server = "RU"
	ServerSG   Server = "SG"
	ServerTR   Server = "TR"
	ServerTW   Server = "TW"
	ServerVN   Server = "VN"
)

// Giá trị Riot tương ứng của 1 server.
// platform: league-v4, summoner-v4. accountRegion: account-v1 (chỉ americas | asia | europe).
// matchRegion: match-v5 (có thêm sea).
type riotRouting struct {
	platform      string
	accountRegion string
	matchRegion   string
}

var serverRiotRoutings = map[Server]riotRouting{
	ServerBR:   {"br1", "americas", "americas"},
	ServerEUNE: {"eun1", "europe", "europe"},
	ServerEUW:  {"euw1", "europe", "europe"},
	ServerJP:   {"jp1", "asia", "asia"},
	ServerKR:   {"kr", "asia", "asia"},
	ServerLAN:  {"la1", "americas", "americas"},
	ServerLAS:  {"la2", "americas", "americas"},
	ServerME:   {"me1", "europe", "europe"},
	ServerNA:   {"na1", "americas", "americas"},
	ServerOCE:  {"oc1", "asia", "sea"},
	ServerRU:   {"ru", "europe", "europe"},
	ServerSG:   {"sg2", "asia", "sea"},
	ServerTR:   {"tr1", "europe", "europe"},
	ServerTW:   {"tw2", "asia", "sea"},
	ServerVN:   {"vn2", "asia", "sea"},
}

// platform viết HOA => Server, dựng ngược từ serverRiotRoutings để không phải khai báo 2 lần.
var serversByPlatform = func() map[string]Server {
	out := make(map[string]Server, len(serverRiotRoutings))
	for s, r := range serverRiotRoutings {
		out[strings.ToUpper(r.platform)] = s
	}
	return out
}()

func (s Server) IsValid() bool {
	_, ok := serverRiotRoutings[s]
	return ok
}

// Giá trị Riot của server. Server không hợp lệ => zero value, nên check IsValid trước.
func (s Server) riot() riotRouting {
	return serverRiotRoutings[s]
}

// platformId của Riot (info.platformId, "VN2") => Server. Không nhận ra => IsNull.
func serverOf(platformId string) shared.Nullable[Server] {
	s, ok := serversByPlatform[strings.ToUpper(platformId)]
	if !ok {
		return shared.Nullable[Server]{IsNull: true}
	}
	return shared.Nullable[Server]{Value: s}
}

// ---------- rank / tier ----------

// Rank trong app = "tier" của Riot.
type Rank string

const (
	RankUnranked    Rank = "UNRANKED"
	RankIron        Rank = "IRON"
	RankBronze      Rank = "BRONZE"
	RankSilver      Rank = "SILVER"
	RankGold        Rank = "GOLD"
	RankPlatinum    Rank = "PLATINUM"
	RankEmerald     Rank = "EMERALD"
	RankDiamond     Rank = "DIAMOND"
	RankMaster      Rank = "MASTER"
	RankGrandmaster Rank = "GRANDMASTER"
	RankChallenger  Rank = "CHALLENGER"
)

// Tier trong app = "rank" của Riot.
type Tier string

const (
	TierI   Tier = "I"
	TierII  Tier = "II"
	TierIII Tier = "III"
	TierIV  Tier = "IV"
)

// Bậc dùng để tính rank power.
var rankLevels = map[Rank]int32{
	RankUnranked:    0,
	RankIron:        1,
	RankBronze:      2,
	RankSilver:      3,
	RankGold:        4,
	RankPlatinum:    5,
	RankEmerald:     6,
	RankDiamond:     7,
	RankMaster:      8,
	RankGrandmaster: 9,
	RankChallenger:  10,
}

var tierLevels = map[Tier]int32{
	TierIV:  0,
	TierIII: 1,
	TierII:  2,
	TierI:   3,
}

const (
	rankPowerPerRank = 400
	rankPowerPerTier = 100
)

// power = bậc rank * 400 + bậc tier * 100 + lp. Unknown => NULL, unranked => 0.
func rankPowerOf(rank shared.Nullable[Rank], tier shared.Nullable[Tier], lp int32) shared.Nullable[int32] {
	if rank.IsNull {
		return shared.Nullable[int32]{IsNull: true}
	}
	if rank.Value == RankUnranked {
		return shared.Nullable[int32]{Value: 0}
	}
	return shared.Nullable[int32]{
		Value: rankLevels[rank.Value]*rankPowerPerRank + tierLevels[tier.Value]*rankPowerPerTier + lp,
	}
}

// Unknown / unranked thì không có tier; Master trở lên tier luôn là IV.
func normalizeTier(rank shared.Nullable[Rank], tier shared.Nullable[Tier]) shared.Nullable[Tier] {
	if rank.IsNull || rank.Value == RankUnranked {
		return shared.Nullable[Tier]{IsNull: true}
	}
	if rankLevels[rank.Value] >= rankLevels[RankMaster] {
		return shared.Nullable[Tier]{Value: TierIV}
	}
	return tier
}

// Trung bình rank power (đã bỏ unknown / UNRANKED) => rank ứng với bậc đó.
// Không participant nào có rank đã biết => UNRANKED (không phải NULL): cột estimated_rank
// là NOT NULL, và analytics lọc theo rank nên UNRANKED bị loại y như unknown.
// Master trở lên LP không giới hạn nên bậc vượt Challenger thì lấy Challenger.
func estimatedRankOf(rankPowers []int32) Rank {
	if len(rankPowers) == 0 {
		return RankUnranked
	}
	var sum int64
	for _, p := range rankPowers {
		sum += int64(p)
	}
	level := int32(sum / int64(len(rankPowers)) / rankPowerPerRank)
	level = max(rankLevels[RankIron], min(level, rankLevels[RankChallenger]))
	for rank, l := range rankLevels {
		if l == level {
			return rank
		}
	}
	return RankUnranked
}

// ---------- game mode ----------

type GameMode string

const (
	GameModeSolo   GameMode = "SOLO"   // queue 420
	GameModeFlex   GameMode = "FLEX"   // queue 440
	GameModeAram   GameMode = "ARAM"   // map 12 (Howling Abyss)
	GameModeNormal GameMode = "NORMAL" // còn lại
)

const (
	riotQueueRankedSolo = 420
	riotQueueRankedFlex = 440
	riotMapHowlingAbyss = 12
)

func gameModeOf(queueId, mapId int) GameMode {
	switch {
	case queueId == riotQueueRankedSolo:
		return GameModeSolo
	case queueId == riotQueueRankedFlex:
		return GameModeFlex
	case mapId == riotMapHowlingAbyss:
		return GameModeAram
	}
	return GameModeNormal
}

// Queue id Riot để lọc match theo mode. ok = false nếu mode không lọc được bằng 1 queue (ARAM, NORMAL).
func riotQueueOf(mode GameMode) (queue int, ok bool) {
	switch mode {
	case GameModeSolo:
		return riotQueueRankedSolo, true
	case GameModeFlex:
		return riotQueueRankedFlex, true
	}
	return 0, false
}

// ---------- position ----------

type Position string

const (
	PositionTop Position = "TOP" // Riot teamPosition TOP
	PositionJgl Position = "JGL" // JUNGLE
	PositionMid Position = "MID" // MIDDLE
	PositionAdc Position = "ADC" // BOTTOM
	PositionSpt Position = "SPT" // UTILITY
	PositionUnk Position = "UNK" // rỗng / giá trị lạ (ARAM, Arena...)
)

// Riot teamPosition => Position. Rỗng / giá trị lạ (ARAM, Arena...) => UNK.
func positionOf(teamPosition string) Position {
	switch teamPosition {
	case "TOP":
		return PositionTop
	case "JUNGLE":
		return PositionJgl
	case "MIDDLE":
		return PositionMid
	case "BOTTOM":
		return PositionAdc
	case "UTILITY":
		return PositionSpt
	}
	return PositionUnk
}

// ---------- item type ----------

type ItemType string

const (
	ItemTypeConsumable ItemType = "CONSUMABLE" // tags có Consumable
	ItemTypeTrinket    ItemType = "TRINKET"    // tags có Trinket
	ItemTypeBoots      ItemType = "BOOTS"      // tags có Boots
	ItemTypeStarter    ItemType = "STARTER"    // không from, không into, còn mua được
	ItemTypeBasic      ItemType = "BASIC"      // không from, có into
	ItemTypeEpic       ItemType = "EPIC"       // có from, có into
	ItemTypeLegendary  ItemType = "LEGENDARY"  // có from, không into
)

// Xét theo thứ tự các ItemType ở trên. ok = false nếu không thuộc loại nào
// (item không mua được, item riêng của tướng...).
func itemTypeOf(it external.DDItem) (t ItemType, ok bool) {
	hasFrom, hasInto := len(it.From) > 0, len(it.Into) > 0
	switch {
	case slices.Contains(it.Tags, "Consumable"):
		return ItemTypeConsumable, true
	case slices.Contains(it.Tags, "Trinket"):
		return ItemTypeTrinket, true
	case slices.Contains(it.Tags, "Boots"):
		return ItemTypeBoots, true
	case !hasFrom && !hasInto && it.Gold.Purchasable:
		return ItemTypeStarter, true
	case !hasFrom && hasInto:
		return ItemTypeBasic, true
	case hasFrom && hasInto:
		return ItemTypeEpic, true
	case hasFrom && !hasInto:
		return ItemTypeLegendary, true
	}
	return "", false
}
