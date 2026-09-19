package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend/src/db"
	"backend/src/external"
	"backend/src/shared"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ---------- enum ----------

// RiotServer = platform của Riot, lưu HOA; gọi API thì dùng chữ thường (vn2, kr...).
type RiotServer string

const (
	ServerBR1  RiotServer = "BR1"
	ServerEUN1 RiotServer = "EUN1"
	ServerEUW1 RiotServer = "EUW1"
	ServerJP1  RiotServer = "JP1"
	ServerKR   RiotServer = "KR"
	ServerLA1  RiotServer = "LA1"
	ServerLA2  RiotServer = "LA2"
	ServerME1  RiotServer = "ME1"
	ServerNA1  RiotServer = "NA1"
	ServerOC1  RiotServer = "OC1"
	ServerRU   RiotServer = "RU"
	ServerSG2  RiotServer = "SG2"
	ServerTR1  RiotServer = "TR1"
	ServerTW2  RiotServer = "TW2"
	ServerVN2  RiotServer = "VN2"
)

// Regional routing của mỗi server.
// account-v1 chỉ có americas | asia | europe; match-v5 có thêm sea.
var serverRegions = map[RiotServer]struct{ account, match string }{
	ServerBR1:  {"americas", "americas"},
	ServerLA1:  {"americas", "americas"},
	ServerLA2:  {"americas", "americas"},
	ServerNA1:  {"americas", "americas"},
	ServerEUN1: {"europe", "europe"},
	ServerEUW1: {"europe", "europe"},
	ServerME1:  {"europe", "europe"},
	ServerRU:   {"europe", "europe"},
	ServerTR1:  {"europe", "europe"},
	ServerJP1:  {"asia", "asia"},
	ServerKR:   {"asia", "asia"},
	ServerOC1:  {"asia", "sea"},
	ServerSG2:  {"asia", "sea"},
	ServerTW2:  {"asia", "sea"},
	ServerVN2:  {"asia", "sea"},
}

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

// FindPlayerByNameAndTag không gọi Riot nếu player vừa update trong khoảng này.
const playerUpdateInterval = 2 * time.Minute

// ---------- entity ----------

type Player struct {
	Id              string                     `json:"id"`
	Server          RiotServer                 `json:"server"`
	Name            string                     `json:"name"`
	Tag             string                     `json:"tag"`
	ProfileIconId   shared.Nullable[int32]     `json:"profileIconId"`
	SummonerLevel   shared.Nullable[int32]     `json:"summonerLevel"`
	SoloRank        shared.Nullable[Rank]      `json:"soloRank"`
	SoloTier        shared.Nullable[Tier]      `json:"soloTier"`
	SoloLp          int32                      `json:"soloLp"`
	SoloRankPower   shared.Nullable[int32]     `json:"soloRankPower"`
	SoloWins        int32                      `json:"soloWins"`
	SoloLosses      int32                      `json:"soloLosses"`
	FlexRank        shared.Nullable[Rank]      `json:"flexRank"`
	FlexTier        shared.Nullable[Tier]      `json:"flexTier"`
	FlexLp          int32                      `json:"flexLp"`
	FlexWins        int32                      `json:"flexWins"`
	FlexLosses      int32                      `json:"flexLosses"`
	LastMatchAt     shared.Nullable[time.Time] `json:"lastMatchAt"`
	MatchesSyncedAt shared.Nullable[time.Time] `json:"matchesSyncedAt"`
	CreatedAt       time.Time                  `json:"createdAt"`
	UpdatedAt       time.Time                  `json:"updatedAt"`
}

func ToPlayer(p db.Player) Player {
	return Player{
		Id:              p.ID,
		Server:          RiotServer(p.Server),
		Name:            p.Name,
		Tag:             p.Tag,
		ProfileIconId:   shared.NullableInt4(p.ProfileIconID),
		SummonerLevel:   shared.NullableInt4(p.SummonerLevel),
		SoloRank:        shared.NullableText[Rank](p.SoloRank),
		SoloTier:        shared.NullableText[Tier](p.SoloTier),
		SoloLp:          p.SoloLp,
		SoloRankPower:   shared.NullableInt4(p.SoloRankPower),
		SoloWins:        p.SoloWins,
		SoloLosses:      p.SoloLosses,
		FlexRank:        shared.NullableText[Rank](p.FlexRank),
		FlexTier:        shared.NullableText[Tier](p.FlexTier),
		FlexLp:          p.FlexLp,
		FlexWins:        p.FlexWins,
		FlexLosses:      p.FlexLosses,
		LastMatchAt:     shared.NullableTime(p.LastMatchAt),
		MatchesSyncedAt: shared.NullableTime(p.MatchesSyncedAt),
		CreatedAt:       p.CreatedAt.Time,
		UpdatedAt:       p.UpdatedAt.Time,
	}
}

// ---------- logic ----------

func (s RiotServer) IsValid() bool {
	_, ok := serverRegions[s]
	return ok
}

// Không tìm thấy => IsNull = true.
func (a *Application) GetPlayerById(ctx context.Context, id string) (shared.Nullable[Player], error) {
	row, err := a.q.GetPlayerById(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.Nullable[Player]{IsNull: true}, nil
	}
	if err != nil {
		return shared.Nullable[Player]{}, err
	}
	return shared.Nullable[Player]{Value: ToPlayer(row)}, nil
}

// Đọc DB trước: player vừa update trong playerUpdateInterval thì trả luôn, không gọi Riot.
// Ngược lại update từ Riot; update lỗi thì trả bản đang có trong DB.
// Không tìm thấy player (Riot 404 và DB không có) => IsNull = true, err = nil.
// Lỗi kỹ thuật (DB lỗi, Riot lỗi mà DB không có...) => err.
func (a *Application) FindPlayerByNameAndTag(ctx context.Context, server RiotServer, name, tag string) (shared.Nullable[Player], error) {
	cached := shared.Nullable[Player]{IsNull: true}
	row, err := a.q.GetPlayerByNameAndTag(ctx, db.GetPlayerByNameAndTagParams{
		NormalizedName: normalizeRiotId(name),
		NormalizedTag:  normalizeRiotId(tag),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return shared.Nullable[Player]{}, err
	}
	if err == nil {
		cached = shared.Nullable[Player]{Value: ToPlayer(row)}
		if time.Since(cached.Value.UpdatedAt) < playerUpdateInterval {
			return cached, nil
		}
	}

	player, updateErr := a.UpdatePlayerByNameAndTag(ctx, server, name, tag)
	if updateErr == nil {
		return shared.Nullable[Player]{Value: player}, nil
	}
	if !cached.IsNull {
		return cached, nil
	}
	if external.IsRiotNotFound(updateErr) {
		return shared.Nullable[Player]{IsNull: true}, nil
	}
	return shared.Nullable[Player]{}, updateErr
}

// Tìm keyword trong name, tag và search_string (đã bỏ dấu). Sắp xếp theo solo rank power giảm dần.
func (a *Application) SearchPlayers(ctx context.Context, keyword string, limit int32) ([]Player, error) {
	keyword = strings.TrimSpace(keyword)
	rows, err := a.q.SearchPlayers(ctx, db.SearchPlayersParams{
		Keyword:           keyword,
		NormalizedKeyword: shared.NormalizeString(keyword),
		Lim:               limit,
	})
	if err != nil {
		return nil, err
	}
	out := []Player{}
	for _, r := range rows {
		out = append(out, ToPlayer(r))
	}
	return out, nil
}

// Lấy account (account-v1), summoner (summoner-v4), rank solo/flex (league-v4) từ Riot rồi upsert vào DB.
// Queue nào không có entry thì là UNRANKED.
func (a *Application) UpdatePlayerByNameAndTag(ctx context.Context, server RiotServer, name, tag string) (Player, error) {
	regions, ok := serverRegions[server]
	if !ok {
		return Player{}, fmt.Errorf("invalid server: %q", server)
	}
	platform := strings.ToLower(string(server))

	account, err := a.riot.GetAccountByRiotId(ctx, regions.account, name, tag)
	if err != nil {
		return Player{}, err
	}
	// summoner và league chỉ cần puuid, không phụ thuộc nhau => gọi song song.
	var (
		summoner    *external.SummonerDto
		entries     []external.LeagueEntryDto
		summonerErr error
		entriesErr  error
		wg          sync.WaitGroup
	)
	wg.Go(func() {
		summoner, summonerErr = a.riot.GetSummonerByPuuid(ctx, platform, account.Puuid)
	})
	wg.Go(func() {
		entries, entriesErr = a.riot.GetLeagueEntriesByPuuid(ctx, platform, account.Puuid)
	})
	wg.Wait()
	err = errors.Join(summonerErr, entriesErr)
	if err != nil {
		return Player{}, err
	}

	solo := queueRankOf(entries, external.QueueRankedSolo)
	flex := queueRankOf(entries, external.QueueRankedFlexSR)
	normalizedName := normalizeRiotId(account.GameName)
	normalizedTag := normalizeRiotId(account.TagLine)

	err = a.q.UpsertPlayer(ctx, db.UpsertPlayerParams{
		ID:             account.Puuid,
		Server:         string(server),
		Name:           account.GameName,
		Tag:            account.TagLine,
		NormalizedName: normalizedName,
		NormalizedTag:  normalizedTag,
		ProfileIconID:  pgtype.Int4{Int32: int32(summoner.ProfileIconId), Valid: true},
		SummonerLevel:  pgtype.Int4{Int32: int32(summoner.SummonerLevel), Valid: true},
		SearchString:   playerSearchString(normalizedName, normalizedTag),

		SoloRank:      shared.PgText(solo.rank),
		SoloTier:      shared.PgText(solo.tier),
		SoloLp:        solo.lp,
		SoloRankPower: shared.PgInt4(rankPowerOf(solo.rank, solo.tier, solo.lp)),
		SoloWins:      solo.wins,
		SoloLosses:    solo.losses,

		FlexRank:   shared.PgText(flex.rank),
		FlexTier:   shared.PgText(flex.tier),
		FlexLp:     flex.lp,
		FlexWins:   flex.wins,
		FlexLosses: flex.losses,
	})
	if err != nil {
		return Player{}, err
	}

	// Vừa upsert xong nên luôn có; không có là lỗi.
	player, err := a.GetPlayerById(ctx, account.Puuid)
	if err != nil {
		return Player{}, err
	}
	if player.IsNull {
		return Player{}, fmt.Errorf("player %s not found after upsert", account.Puuid)
	}
	return player.Value, nil
}

// ---------- private ----------

type queueRank struct {
	rank   shared.Nullable[Rank]
	tier   shared.Nullable[Tier]
	lp     int32
	wins   int32
	losses int32
}

// Không có entry cho queue => UNRANKED. Tier đã qua normalizeTier.
func queueRankOf(entries []external.LeagueEntryDto, queue external.QueueType) queueRank {
	for _, e := range entries {
		if e.QueueType != string(queue) {
			continue
		}
		rank := shared.Nullable[Rank]{Value: Rank(e.Tier)} // Riot "tier" = rank trong app
		tier := shared.Nullable[Tier]{Value: Tier(e.Rank)} // Riot "rank" = tier trong app
		return queueRank{
			rank:   rank,
			tier:   normalizeTier(rank, tier),
			lp:     int32(e.LeaguePoints),
			wins:   int32(e.Wins),
			losses: int32(e.Losses),
		}
	}
	return queueRank{
		rank: shared.Nullable[Rank]{Value: RankUnranked},
		tier: shared.Nullable[Tier]{IsNull: true},
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

// Chữ thường + trim, giữ dấu. Dùng cho normalized_name / normalized_tag để find chính xác.
func normalizeRiotId(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Nhận normalized_name / normalized_tag, bỏ dấu thêm để search.
func playerSearchString(normalizedName, normalizedTag string) string {
	return shared.NormalizeString(normalizedName) + "#" + shared.NormalizeString(normalizedTag)
}
