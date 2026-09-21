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

// FindPlayerByPlayerInfo không gọi Riot nếu player vừa update trong khoảng này.
const playerUpdateInterval = 2 * time.Minute

// ---------- entity ----------

type Player struct {
	Id            string                 `json:"id"`
	Server        Server                 `json:"server"`
	Name          string                 `json:"name"`
	Tag           string                 `json:"tag"`
	ProfileIconId shared.Nullable[int32] `json:"profileIconId"`
	Level         shared.Nullable[int32] `json:"level"`
	SoloRank      shared.Nullable[Rank]  `json:"soloRank"`
	SoloTier      shared.Nullable[Tier]  `json:"soloTier"`
	SoloLp        int32                  `json:"soloLp"`
	SoloRankPower shared.Nullable[int32] `json:"soloRankPower"`
	SoloWins      int32                  `json:"soloWins"`
	SoloLosses    int32                  `json:"soloLosses"`
	FlexRank      shared.Nullable[Rank]  `json:"flexRank"`
	FlexTier      shared.Nullable[Tier]  `json:"flexTier"`
	FlexLp        int32                  `json:"flexLp"`
	FlexWins      int32                  `json:"flexWins"`
	FlexLosses    int32                  `json:"flexLosses"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

func ToPlayer(p db.LolPlayer) Player {
	return Player{
		Id:            p.ID,
		Server:        Server(p.Server),
		Name:          p.Name,
		Tag:           p.Tag,
		ProfileIconId: shared.NullableInt4(p.ProfileIconID),
		Level:         shared.NullableInt4(p.Level),
		SoloRank:      shared.NullableText[Rank](p.SoloRank),
		SoloTier:      shared.NullableText[Tier](p.SoloTier),
		SoloLp:        p.SoloLp,
		SoloRankPower: shared.NullableInt4(p.SoloRankPower),
		SoloWins:      p.SoloWins,
		SoloLosses:    p.SoloLosses,
		FlexRank:      shared.NullableText[Rank](p.FlexRank),
		FlexTier:      shared.NullableText[Tier](p.FlexTier),
		FlexLp:        p.FlexLp,
		FlexWins:      p.FlexWins,
		FlexLosses:    p.FlexLosses,
		CreatedAt:     p.CreatedAt.Time,
		UpdatedAt:     p.UpdatedAt.Time,
	}
}

// ---------- logic ----------

// Không tìm thấy => IsNull = true.
func (a *Application) GetPlayerById(ctx context.Context, id string) (Player, error) {
	row, err := a.q.GetPlayerById(ctx, id)
	if err != nil {
		return Player{}, err
	}
	return ToPlayer(row), nil
}

// Đọc DB trước: player vừa update trong playerUpdateInterval thì trả luôn, không gọi Riot.
// Ngược lại update từ Riot; update lỗi thì trả bản đang có trong DB.
// Không tìm thấy player (Riot 404 và DB không có) => IsNull = true, err = nil.
// Lỗi kỹ thuật (DB lỗi, Riot lỗi mà DB không có...) => err.
func (a *Application) FindPlayerByPlayerInfo(ctx context.Context, server Server, name, tag string) (shared.Nullable[Player], error) {
	cached := shared.Nullable[Player]{IsNull: true}
	row, err := a.q.GetPlayerByServerNameAndTag(ctx, db.GetPlayerByServerNameAndTagParams{
		Server:         string(server),
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

	player, updateErr := a.UpdatePlayerByPlayerInfo(ctx, server, name, tag)
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
func (a *Application) UpdatePlayerByPlayerInfo(ctx context.Context, server Server, name, tag string) (Player, error) {
	if !server.IsValid() {
		return Player{}, fmt.Errorf("invalid server: %q", server)
	}

	account, err := a.riot.GetAccountByRiotId(ctx, server.riot().accountRegion, name, tag)
	if err != nil {
		return Player{}, err
	}
	return a.updatePlayerFromAccount(ctx, server, account)
}

// Như UpdatePlayerByPlayerInfo nhưng vào thẳng bằng puuid: chỗ nào có sẵn puuid (crawler) thì
// không phải đi vòng qua name/tag, đỡ 1 call account-v1.
func (a *Application) UpdatePlayerByPuuid(ctx context.Context, server Server, puuid string) (Player, error) {
	if !server.IsValid() {
		return Player{}, fmt.Errorf("invalid server: %q", server)
	}

	account, err := a.riot.GetAccountByPuuid(ctx, server.riot().accountRegion, puuid)
	if err != nil {
		return Player{}, err
	}
	return a.updatePlayerFromAccount(ctx, server, account)
}

// ---------- private ----------

// Lấy summoner (summoner-v4) và rank solo/flex (league-v4) của account rồi upsert vào DB.
// Queue nào không có entry thì là UNRANKED.
func (a *Application) updatePlayerFromAccount(ctx context.Context, server Server, account *external.AccountDto) (Player, error) {
	routing := server.riot()
	// summoner và league chỉ cần puuid, không phụ thuộc nhau => gọi song song.
	var (
		summoner    *external.SummonerDto
		entries     []external.LeagueEntryDto
		summonerErr error
		entriesErr  error
		wg          sync.WaitGroup
	)
	wg.Go(func() {
		summoner, summonerErr = a.riot.GetSummonerByPuuid(ctx, routing.platform, account.Puuid)
	})
	wg.Go(func() {
		entries, entriesErr = a.riot.GetLeagueEntriesByPuuid(ctx, routing.platform, account.Puuid)
	})
	wg.Wait()
	err := errors.Join(summonerErr, entriesErr)
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
		Level:          pgtype.Int4{Int32: int32(summoner.SummonerLevel), Valid: true},
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
	return player, nil
}

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

// Chữ thường + trim, giữ dấu. Dùng cho normalized_name / normalized_tag để find chính xác.
func normalizeRiotId(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Nhận normalized_name / normalized_tag, bỏ dấu thêm để search.
func playerSearchString(normalizedName, normalizedTag string) string {
	return shared.NormalizeString(normalizedName) + "#" + shared.NormalizeString(normalizedTag)
}
