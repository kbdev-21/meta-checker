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

// server đặc biệt của meta: gộp mọi server.
const MetaServerGlobal = "GLOBAL"

// server hợp lệ cho meta = GLOBAL hoặc một Server của app.
func IsValidMetaServer(server string) bool {
	return server == MetaServerGlobal || Server(server).IsValid()
}

// ---------- entity ----------

// Các phần tử build lưu trong JSONB, mỗi phần tử kèm games/wins để đứng độc lập.
type SpellComboStat struct {
	Spell1Id int32 `json:"spell1Id"`
	Spell2Id int32 `json:"spell2Id"`
	Games    int32 `json:"games"`
	Wins     int32 `json:"wins"`
}

type RuneStat struct {
	RunePrimaryStyle int32   `json:"runePrimaryStyle"`
	RuneSubStyle     int32   `json:"runeSubStyle"`
	KeyRune          int32   `json:"keyRune"`
	Runes            []int32 `json:"runes"`
	StatRunes        []int32 `json:"statRunes"`
	Games            int32   `json:"games"`
	Wins             int32   `json:"wins"`
}

type ItemStat struct {
	ItemId int32 `json:"itemId"`
	Games  int32 `json:"games"`
	Wins   int32 `json:"wins"`
}

type MatchupStat struct {
	OpponentChampionId int32 `json:"opponentChampionId"`
	Games              int32 `json:"games"`
	Wins               int32 `json:"wins"`
}

// ChampionStat nhúng luôn patch/server/rankBucket của meta để có thể đứng độc lập
// (tách khỏi Meta vẫn biết mình thuộc lát cắt nào).
type ChampionStat struct {
	Patch      string     `json:"patch"`
	Server     string     `json:"server"`
	RankBucket RankBucket `json:"rankBucket"`

	Position     Position `json:"position"`
	ChampionId   int32    `json:"championId"`
	ChampionSlug string   `json:"championSlug"`

	Games    int32   `json:"games"`
	Wins     int32   `json:"wins"`
	WinRate  float64 `json:"winRate"`
	PickRate float64 `json:"pickRate"`
	Bans     int32   `json:"bans"`
	BanRate  float64 `json:"banRate"`
	Power    int32   `json:"power"` // tính lúc đọc từ winRate + pickRate, không lưu DB

	AvgKills       float64 `json:"avgKills"`
	AvgDeaths      float64 `json:"avgDeaths"`
	AvgAssists     float64 `json:"avgAssists"`
	AvgKda         float64 `json:"avgKda"`
	AvgKp          float64 `json:"avgKp"`
	AvgCsPerMin    float64 `json:"avgCsPerMin"`
	AvgGoldPerMin  float64 `json:"avgGoldPerMin"`
	AvgDmgPerMin   float64 `json:"avgDmgPerMin"`
	AvgPhysicalDmg float64 `json:"avgPhysicalDmg"`
	AvgMagicDmg    float64 `json:"avgMagicDmg"`
	AvgTrueDmg     float64 `json:"avgTrueDmg"`
	AvgPenta       float64 `json:"avgPenta"`
	AvgSoloKills   float64 `json:"avgSoloKills"`
	AvgPerfScore   float64 `json:"avgPerfScore"`

	BestSpellCombos    []SpellComboStat `json:"bestSpellCombos"`
	BestRunes          []RuneStat       `json:"bestRunes"`
	BestLegendaryItems []ItemStat       `json:"bestLegendaryItems"`
	BestBootItems      []ItemStat       `json:"bestBootItems"`
	Matchups           []MatchupStat    `json:"matchups"`
}

type Meta struct {
	Id            string         `json:"id"`
	Patch         string         `json:"patch"`
	Server        string         `json:"server"`
	RankBucket    RankBucket     `json:"rankBucket"`
	TotalMatches  int32          `json:"totalMatches"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	ChampionStats []ChampionStat `json:"championStats"`
}

func ToChampionStat(m db.LolMeta, r db.GetChampionStatsByMetaRow) ChampionStat {
	return ChampionStat{
		Patch:      m.Patch,
		Server:     m.Server,
		RankBucket: RankBucket(m.RankBucket),

		Position:     Position(r.Position),
		ChampionId:   r.ChampionID,
		ChampionSlug: r.ChampionSlug,

		Games:    r.Games,
		Wins:     r.Wins,
		WinRate:  r.WinRate,
		PickRate: r.PickRate,
		Bans:     r.Bans,
		BanRate:  r.BanRate,
		Power:    champPowerOf(r.WinRate, r.PickRate),

		AvgKills:       r.AvgKills,
		AvgDeaths:      r.AvgDeaths,
		AvgAssists:     r.AvgAssists,
		AvgKda:         r.AvgKda,
		AvgKp:          r.AvgKp,
		AvgCsPerMin:    r.AvgCsPerMin,
		AvgGoldPerMin:  r.AvgGoldPerMin,
		AvgDmgPerMin:   r.AvgDmgPerMin,
		AvgPhysicalDmg: r.AvgPhysicalDmg,
		AvgMagicDmg:    r.AvgMagicDmg,
		AvgTrueDmg:     r.AvgTrueDmg,
		AvgPenta:       r.AvgPenta,
		AvgSoloKills:   r.AvgSoloKills,
		AvgPerfScore:   r.AvgPerfScore,

		BestSpellCombos:    unmarshalBuild[SpellComboStat](r.BestSpellCombos),
		BestRunes:          unmarshalBuild[RuneStat](r.BestRunes),
		BestLegendaryItems: unmarshalBuild[ItemStat](r.BestLegendaryItems),
		BestBootItems:      unmarshalBuild[ItemStat](r.BestBootItems),
		Matchups:           unmarshalBuild[MatchupStat](r.Matchups),
	}
}

func ToMeta(m db.LolMeta, stats []ChampionStat) Meta {
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

// Slice test hiện tại. Mở rộng thêm server / bucket ở đây khi cần.
var (
	analyticsServers = []string{MetaServerGlobal}
	analyticsBuckets = []RankBucket{RankBucketMasterPlus}
)

const (
	analyticsMinDurationSec = 300 // bỏ trận quá ngắn
	analyticsTopN           = 5   // số phần tử mỗi nhóm build tốt nhất
)

// Tổng hợp lại toàn bộ meta cho patch hiện tại. Idempotent: xóa sạch children rồi dựng lại.
func (a *Application) UpdateAnalytics(ctx context.Context) error {
	version, err := a.ddragon.GetCurrentVersion(ctx)
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
	version, err := a.ddragon.GetCurrentVersion(ctx)
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
	stats := make([]ChampionStat, 0, len(rows))
	for _, r := range rows {
		stats = append(stats, ToChampionStat(m, r))
	}
	return shared.Nullable[Meta]{Value: ToMeta(m, stats)}, nil
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
		MetaID: metaId, TopN: analyticsTopN, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsRunes(ctx, db.RefreshChampionStatsRunesParams{
		MetaID: metaId, TopN: analyticsTopN, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsLegendaryItems(ctx, db.RefreshChampionStatsLegendaryItemsParams{
		MetaID: metaId, TopN: analyticsTopN, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsBootItems(ctx, db.RefreshChampionStatsBootItemsParams{
		MetaID: metaId, TopN: analyticsTopN, Patch: patch,
		MinDuration: analyticsMinDurationSec, Ranks: ranks, Server: server,
	})
	if err != nil {
		return err
	}
	err = q.RefreshChampionStatsMatchups(ctx, db.RefreshChampionStatsMatchupsParams{
		MetaID: metaId, TopN: analyticsTopN, Patch: patch,
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

// Các estimated_rank thuộc 1 bucket.
func bucketRanks(b RankBucket) []string {
	switch b {
	case RankBucketMasterPlus:
		return []string{string(RankMaster), string(RankGrandmaster), string(RankChallenger)}
	}
	return nil
}

// power = winScoreRatio*60 + pickScoreRatio*40, mỗi ratio clamp về [0,1]. Cắt về int như bản Kotlin.
func champPowerOf(winRate, pickRate float64) int32 {
	const winWeight, pickWeight = 60.0, 40.0
	const minWinRate, maxWinRate = 0.46, 0.54
	const minPickRate, maxPickRate = 0.0, 0.075

	winScoreRatio := clamp01((winRate - minWinRate) / (maxWinRate - minWinRate))
	pickScoreRatio := clamp01((pickRate - minPickRate) / (maxPickRate - minPickRate))
	return int32(winScoreRatio*winWeight + pickScoreRatio*pickWeight)
}

func clamp01(x float64) float64 {
	return max(0.0, min(1.0, x))
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
