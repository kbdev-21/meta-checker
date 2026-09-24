package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/src/db"
	"backend/src/shared"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ---------- enum ----------

// Nhóm rank để tổng hợp meta. MASTER_PLUS = estimated_rank thuộc MASTER | GRANDMASTER | CHALLENGER.
type RankBucket string

const (
	RankBucketMasterPlus RankBucket = "MASTER_PLUS"
)

func (b RankBucket) IsValid() bool {
	return b == RankBucketMasterPlus
}

// Bậc sức mạnh của tướng trong meta, suy từ TierScore (thang 100 chia đều 5 bậc).
type ChampionTier string

const (
	ChampionTierS ChampionTier = "S"
	ChampionTierA ChampionTier = "A"
	ChampionTierB ChampionTier = "B"
	ChampionTierC ChampionTier = "C"
	ChampionTierD ChampionTier = "D"
)

// server đặc biệt của meta: gộp mọi server.
const MetaServerGlobal = "GLOBAL"

// server hợp lệ cho meta = GLOBAL hoặc một Server của app.
func IsValidMetaServer(server string) bool {
	return server == MetaServerGlobal || Server(server).IsValid()
}

// ---------- entity ----------

// Số đếm chung của mọi phần tử build trong JSONB, để mỗi phần tử đứng độc lập.
// Nhúng không kèm json tag nên key "games" / "wins" nằm phẳng cạnh các key còn lại.
type GameStat struct {
	Games int32 `json:"games"`
	Wins  int32 `json:"wins"`
}

type SpellComboStat struct {
	Spell1Id int32 `json:"spell1Id"`
	Spell2Id int32 `json:"spell2Id"`
	GameStat
}

type RuneStat struct {
	RunePrimaryStyle int32   `json:"runePrimaryStyle"`
	RuneSubStyle     int32   `json:"runeSubStyle"`
	KeyRune          int32   `json:"keyRune"`
	Runes            []int32 `json:"runes"`
	StatRunes        []int32 `json:"statRunes"`
	GameStat
}

type ItemStat struct {
	ItemId int32 `json:"itemId"`
	GameStat
}

type MatchupStat struct {
	OpponentChampionId int32 `json:"opponentChampionId"`
	GameStat
}

// Một bộ item: starter set (đã sort) hoặc 3 đồ legendary đầu (giữ thứ tự mua).
type ItemSetStat struct {
	ItemIds []int32 `json:"itemIds"`
	GameStat
}

// Thứ tự lên skill, mỗi phần tử là skillSlot: 1 = Q, 2 = W, 3 = E, 4 = R.
type SkillOrderStat struct {
	Skills []int32 `json:"skills"`
	GameStat
}

// ChampionStatSummary nhúng luôn patch/server/rankBucket của meta để có thể đứng độc lập
// (tách khỏi Meta vẫn biết mình thuộc lát cắt nào). Không kèm các nhóm build: tier list trả
// hàng trăm dòng nên chúng sẽ chiếm ~85% payload.
type ChampionStatSummary struct {
	Patch      string     `json:"patch"`
	Server     string     `json:"server"`
	RankBucket RankBucket `json:"rankBucket"`

	Position     Position `json:"position"`
	ChampionId   int32    `json:"championId"`
	ChampionSlug string   `json:"championSlug"`

	Games     int32        `json:"games"`
	Wins      int32        `json:"wins"`
	WinRate   float64      `json:"winRate"`
	PickRate  float64      `json:"pickRate"`
	Bans      int32        `json:"bans"`
	BanRate   float64      `json:"banRate"`
	TierScore int32        `json:"tierScore"` // tierScore + tier tính lúc đọc từ winRate + pickRate, không lưu DB
	Tier      ChampionTier `json:"tier"`

	AvgKills          float64 `json:"avgKills"`
	AvgDeaths         float64 `json:"avgDeaths"`
	AvgAssists        float64 `json:"avgAssists"`
	AvgKda            float64 `json:"avgKda"`
	AvgKp             float64 `json:"avgKp"`
	AvgCsPerMin       float64 `json:"avgCsPerMin"`
	AvgGoldPerMin     float64 `json:"avgGoldPerMin"`
	AvgDmgPerMin      float64 `json:"avgDmgPerMin"`
	AvgDmgTakenPerMin float64 `json:"avgDmgTakenPerMin"`
	AvgCcPerMin       float64 `json:"avgCcPerMin"`
	AvgPhysicalDmg    float64 `json:"avgPhysicalDmg"`
	AvgMagicDmg       float64 `json:"avgMagicDmg"`
	AvgTrueDmg        float64 `json:"avgTrueDmg"`
	AvgPenta          float64 `json:"avgPenta"`
	AvgSoloKills      float64 `json:"avgSoloKills"`
	AvgPerfScore      float64 `json:"avgPerfScore"`

	BestMatchUps []MatchupStat `json:"bestMatchUps"`
}

// ChampionStat = summary + build, cho endpoint đọc 1 tướng. Struct nhúng nên JSON vẫn
// phẳng: key y hệt ChampionStatSummary cộng thêm các key build.
type ChampionStat struct {
	ChampionStatSummary

	BestSpellCombos      []SpellComboStat `json:"bestSpellCombos"`
	BestRunes            []RuneStat       `json:"bestRunes"`
	BestLegendaryItems   []ItemStat       `json:"bestLegendaryItems"`
	BestBootItems        []ItemStat       `json:"bestBootItems"`
	BestStarterSets      []ItemSetStat    `json:"bestStarterSets"`
	BestSkillsLeveled    []SkillOrderStat `json:"bestSkillsLeveled"`
	BestFirstLegendItems []ItemStat       `json:"bestFirstLegendItems"`
	BestFirstThreeItems  []ItemSetStat    `json:"bestFirstThreeItems"`
}

type Meta struct {
	Id            string                `json:"id"`
	Patch         string                `json:"patch"`
	Server        string                `json:"server"`
	RankBucket    RankBucket            `json:"rankBucket"`
	TotalMatches  int32                 `json:"totalMatches"`
	UpdatedAt     time.Time             `json:"updatedAt"`
	ChampionStats []ChampionStatSummary `json:"championStats"`
}

func ToChampionStatSummary(m db.LolMeta, r db.GetChampionStatsByMetaRow) ChampionStatSummary {
	tierScore := tierScoreOf(r.WinRate, r.PickRate)

	return ChampionStatSummary{
		Patch:      m.Patch,
		Server:     m.Server,
		RankBucket: RankBucket(m.RankBucket),

		Position:     Position(r.Position),
		ChampionId:   r.ChampionID,
		ChampionSlug: r.ChampionSlug,

		Games:     r.Games,
		Wins:      r.Wins,
		WinRate:   r.WinRate,
		PickRate:  r.PickRate,
		Bans:      r.Bans,
		BanRate:   r.BanRate,
		TierScore: tierScore,
		Tier:      championTierOf(tierScore),

		AvgKills:          r.AvgKills,
		AvgDeaths:         r.AvgDeaths,
		AvgAssists:        r.AvgAssists,
		AvgKda:            r.AvgKda,
		AvgKp:             r.AvgKp,
		AvgCsPerMin:       r.AvgCsPerMin,
		AvgGoldPerMin:     r.AvgGoldPerMin,
		AvgDmgPerMin:      r.AvgDmgPerMin,
		AvgDmgTakenPerMin: r.AvgDmgTakenPerMin,
		AvgCcPerMin:       r.AvgCcPerMin,
		AvgPhysicalDmg:    r.AvgPhysicalDmg,
		AvgMagicDmg:       r.AvgMagicDmg,
		AvgTrueDmg:        r.AvgTrueDmg,
		AvgPenta:          r.AvgPenta,
		AvgSoloKills:      r.AvgSoloKills,
		AvgPerfScore:      r.AvgPerfScore,

		BestMatchUps: unmarshalBuild[MatchupStat](r.BestMatchUps),
	}
}

// Row của query 1 tướng là type khác (có thêm các cột build) nên sqlc không dùng chung
// struct được; phần scalar vì vậy map lại ở đây, phải sửa kèm ToChampionStatSummary.
func ToChampionStat(m db.LolMeta, r db.GetChampionStatsByMetaAndChampIdRow) ChampionStat {
	tierScore := tierScoreOf(r.WinRate, r.PickRate)

	return ChampionStat{
		ChampionStatSummary: ChampionStatSummary{
			Patch:      m.Patch,
			Server:     m.Server,
			RankBucket: RankBucket(m.RankBucket),

			Position:     Position(r.Position),
			ChampionId:   r.ChampionID,
			ChampionSlug: r.ChampionSlug,

			Games:     r.Games,
			Wins:      r.Wins,
			WinRate:   r.WinRate,
			PickRate:  r.PickRate,
			Bans:      r.Bans,
			BanRate:   r.BanRate,
			TierScore: tierScore,
			Tier:      championTierOf(tierScore),

			AvgKills:          r.AvgKills,
			AvgDeaths:         r.AvgDeaths,
			AvgAssists:        r.AvgAssists,
			AvgKda:            r.AvgKda,
			AvgKp:             r.AvgKp,
			AvgCsPerMin:       r.AvgCsPerMin,
			AvgGoldPerMin:     r.AvgGoldPerMin,
			AvgDmgPerMin:      r.AvgDmgPerMin,
			AvgDmgTakenPerMin: r.AvgDmgTakenPerMin,
			AvgCcPerMin:       r.AvgCcPerMin,
			AvgPhysicalDmg:    r.AvgPhysicalDmg,
			AvgMagicDmg:       r.AvgMagicDmg,
			AvgTrueDmg:        r.AvgTrueDmg,
			AvgPenta:          r.AvgPenta,
			AvgSoloKills:      r.AvgSoloKills,
			AvgPerfScore:      r.AvgPerfScore,

			BestMatchUps: unmarshalBuild[MatchupStat](r.BestMatchUps),
		},

		BestSpellCombos:      unmarshalBuild[SpellComboStat](r.BestSpellCombos),
		BestRunes:            unmarshalBuild[RuneStat](r.BestRunes),
		BestLegendaryItems:   unmarshalBuild[ItemStat](r.BestLegendaryItems),
		BestBootItems:        unmarshalBuild[ItemStat](r.BestBootItems),
		BestStarterSets:      unmarshalBuild[ItemSetStat](r.BestStarterSets),
		BestSkillsLeveled:    unmarshalBuild[SkillOrderStat](r.BestSkillsLeveled),
		BestFirstLegendItems: unmarshalBuild[ItemStat](r.BestFirstLegendItems),
		BestFirstThreeItems:  unmarshalBuild[ItemSetStat](r.BestFirstThreeItems),
	}
}

func ToMeta(m db.LolMeta, stats []ChampionStatSummary) Meta {
	return Meta{
		Id:            uuidString(m.ID),
		Patch:         m.Patch,
		Server:        m.Server,
		RankBucket:    RankBucket(m.RankBucket),
		TotalMatches:  m.TotalMatches,
		UpdatedAt:     m.UpdatedAt.Time,
		ChampionStats: stats,
	}
}

// ---------- logic ----------

// GLOBAL + mọi server đang crawl (suy từ Crawlers để khỏi khai báo trùng).
// Thêm bucket ở đây khi cần.
var (
	analyticsServers = func() []string {
		out := []string{MetaServerGlobal}
		for _, c := range Crawlers {
			out = append(out, string(c.Server))
		}
		return out
	}()
	analyticsBuckets = []RankBucket{RankBucketMasterPlus}
)

const (
	analyticsMinDurationSec = 300 // bỏ trận quá ngắn
)

// Tổng hợp lại toàn bộ meta cho patch hiện tại. Idempotent: xóa sạch children rồi dựng lại.
func (a *Application) UpdateAnalytics(ctx context.Context) error {
	version, err := a.ddragon.FetchCurrentVersion(ctx)
	if err != nil {
		return err
	}
	patch := shared.PatchOf(version)

	for _, bucket := range analyticsBuckets {
		for _, server := range analyticsServers {
			err = a.refreshMeta(ctx, patch, server, bucket)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Trả meta của patch hiện tại kèm tất cả champion stat. Chưa tổng hợp lần nào => IsNull.
func (a *Application) GetAnalytics(ctx context.Context, server string, bucket RankBucket) (shared.Nullable[Meta], error) {
	version, err := a.ddragon.FetchCurrentVersion(ctx)
	if err != nil {
		return shared.Nullable[Meta]{}, err
	}
	patch := shared.PatchOf(version)

	m, err := a.q.GetMeta(ctx, db.GetMetaParams{Patch: patch, Server: server, RankBucket: string(bucket)})
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.Nullable[Meta]{IsNull: true}, nil
	}
	if err != nil {
		return shared.Nullable[Meta]{}, err
	}

	rows, err := a.q.GetChampionStatsByMeta(ctx, m.ID)
	if err != nil {
		return shared.Nullable[Meta]{}, err
	}
	stats := make([]ChampionStatSummary, 0, len(rows))
	for _, r := range rows {
		stats = append(stats, ToChampionStatSummary(m, r))
	}
	return shared.Nullable[Meta]{Value: ToMeta(m, stats)}, nil
}

// Trả mọi position của 1 champion trong lát cắt của patch hiện tại.
// Chưa tổng hợp meta => IsNull; champion không có dòng nào => list rỗng.
func (a *Application) GetChampionStats(ctx context.Context, server string, bucket RankBucket, championId int32) (shared.Nullable[[]ChampionStat], error) {
	version, err := a.ddragon.FetchCurrentVersion(ctx)
	if err != nil {
		return shared.Nullable[[]ChampionStat]{}, err
	}
	patch := shared.PatchOf(version)

	m, err := a.q.GetMeta(ctx, db.GetMetaParams{Patch: patch, Server: server, RankBucket: string(bucket)})
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.Nullable[[]ChampionStat]{IsNull: true}, nil
	}
	if err != nil {
		return shared.Nullable[[]ChampionStat]{}, err
	}

	rows, err := a.q.GetChampionStatsByMetaAndChampId(ctx, db.GetChampionStatsByMetaAndChampIdParams{MetaID: m.ID, ChampionID: championId})
	if err != nil {
		return shared.Nullable[[]ChampionStat]{}, err
	}
	stats := make([]ChampionStat, 0, len(rows))
	for _, r := range rows {
		stats = append(stats, ToChampionStat(m, r))
	}
	return shared.Nullable[[]ChampionStat]{Value: stats}, nil
}

// ---------- private ----------

// Tổng hợp 1 lát cắt (patch, server, bucket) trong 1 transaction: upsert meta,
// xóa children cũ, insert scalar, điền JSONB build, insert ban.
func (a *Application) refreshMeta(ctx context.Context, patch, server string, bucket RankBucket) error {
	ranks := bucketRanks(bucket)

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)

	total, err := q.CountSliceMatches(ctx, db.CountSliceMatchesParams{
		Patch: patch, MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}

	metaId, err := q.UpsertMeta(ctx, db.UpsertMetaParams{
		Patch: patch, Server: server, RankBucket: string(bucket), TotalMatches: int32(total),
	})
	if err != nil {
		return err
	}

	err = q.DeleteChampionStatsByMeta(ctx, metaId)
	if err != nil {
		return err
	}
	err = q.DeleteChampionBansByMeta(ctx, metaId)
	if err != nil {
		return err
	}

	err = q.RefreshChampionStatsScalars(ctx, db.RefreshChampionStatsScalarsParams{
		MetaID: metaId, TotalMatches: int32(total), Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}

	err = q.RefreshChampionStatsSpellCombos(ctx, db.RefreshChampionStatsSpellCombosParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsRunes(ctx, db.RefreshChampionStatsRunesParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	transformedIds, baseIds := transformedItemPairs()
	err = q.RefreshChampionStatsLegendaryItems(ctx, db.RefreshChampionStatsLegendaryItemsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
		TransformedIds: transformedIds, BaseIds: baseIds,
	})
	if err != nil {
		return err
	}
	items, err := q.ListItems(ctx)
	if err != nil {
		return err
	}
	tierThreeIds, tierTwoIds := bootTierTwoPairs(items)
	err = q.RefreshChampionStatsBootItems(ctx, db.RefreshChampionStatsBootItemsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
		TransformedIds: tierThreeIds, BaseIds: tierTwoIds,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsMatchups(ctx, db.RefreshChampionStatsMatchupsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsStarterSets(ctx, db.RefreshChampionStatsStarterSetsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsSkillsLeveled(ctx, db.RefreshChampionStatsSkillsLeveledParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsFirstLegendItems(ctx, db.RefreshChampionStatsFirstLegendItemsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsFirstThreeItems(ctx, db.RefreshChampionStatsFirstThreeItemsParams{
		MetaID: metaId, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}

	err = q.RefreshChampionBans(ctx, db.RefreshChampionBansParams{
		MetaID: metaId, TotalMatches: int32(total), Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// transformedItemBases tách thành 2 mảng song song để truyền vào SQL: transformedIds[i] => baseIds[i].
func transformedItemPairs() (transformedIds, baseIds []int32) {
	for transformed, base := range transformedItemBases {
		transformedIds = append(transformedIds, transformed)
		baseIds = append(baseIds, base)
	}
	return transformedIds, baseIds
}

// Giày tier 3 => giày tier 2 tương ứng, tách thành 2 mảng song song như transformedItemPairs.
// Tier 2 = BOOTS chỉ ghép từ giày thường (1001 là BASIC); tier 3 = BOOTS có thành phần là BOOTS.
// Đi ngược chuỗi from tới tier 2 vì có giày qua nhiều bậc (Forever Forward 3176 => Synchronized
// Souls 3013 => Symbiotic Soles 3010). Suy từ lol_items để Riot thêm giày mới thì tự nhận.
func bootTierTwoPairs(items []db.LolItem) (tierThreeIds, tierTwoIds []int32) {
	boots := map[int32]db.LolItem{}
	for _, it := range items {
		if ItemType(it.Type) == ItemTypeBoots {
			boots[it.ID] = it
		}
	}
	// Thành phần là BOOTS của 1 giày; không có => giày đó là tier 2.
	bootsFromOf := func(it db.LolItem) (db.LolItem, bool) {
		for _, id := range it.FromItems {
			if from, ok := boots[id]; ok {
				return from, true
			}
		}
		return db.LolItem{}, false
	}
	for _, it := range boots {
		base, ok := bootsFromOf(it)
		if !ok {
			continue
		}
		for {
			next, ok := bootsFromOf(base)
			if !ok {
				break
			}
			base = next
		}
		tierThreeIds = append(tierThreeIds, it.ID)
		tierTwoIds = append(tierTwoIds, base.ID)
	}
	return tierThreeIds, tierTwoIds
}

// Các estimated_rank thuộc 1 bucket.
func bucketRanks(b RankBucket) []string {
	switch b {
	case RankBucketMasterPlus:
		return []string{string(RankMaster), string(RankGrandmaster), string(RankChallenger)}
	}
	return nil
}

// tierScore = winScoreRatio*60 + pickScoreRatio*40, mỗi ratio clamp về [0,1]. Cắt về int như bản Kotlin.
func tierScoreOf(winRate, pickRate float64) int32 {
	const winWeight, pickWeight = 65.0, 35.0
	const minWinRate, maxWinRate = 0.46, 0.54
	const minPickRate, maxPickRate = 0.0, 0.05

	winScoreRatio := clamp01((winRate - minWinRate) / (maxWinRate - minWinRate))
	pickScoreRatio := clamp01((pickRate - minPickRate) / (maxPickRate - minPickRate))
	return int32(winScoreRatio*winWeight + pickScoreRatio*pickWeight)
}

func clamp01(x float64) float64 {
	return max(0.0, min(1.0, x))
}

// Thang 100 chia đều 5 bậc: S 80-100, A 60-79, B 40-59, C 20-39, D 0-19.
func championTierOf(tierScore int32) ChampionTier {
	switch {
	case tierScore >= 80:
		return ChampionTierS
	case tierScore >= 60:
		return ChampionTierA
	case tierScore >= 40:
		return ChampionTierB
	case tierScore >= 20:
		return ChampionTierC
	}
	return ChampionTierD
}

// JSONB build luôn là mảng hợp lệ (cột DEFAULT '[]'); lỗi bất thường => mảng rỗng.
func unmarshalBuild[T any](b []byte) []T {
	out := []T{}
	if len(b) == 0 {
		return out
	}
	_ = json.Unmarshal(b, &out)
	return out
}

func uuidString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
