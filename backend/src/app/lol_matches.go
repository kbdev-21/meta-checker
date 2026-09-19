package app

import (
	"context"
	"fmt"
	"time"

	"backend/src/db"
	"backend/src/external"
	"backend/src/shared"

	"github.com/jackc/pgx/v5/pgtype"
)

// ---------- enum ----------

type LolGameMode string

const (
	LolGameModeSolo   LolGameMode = "SOLO"   // queue 420
	LolGameModeFlex   LolGameMode = "FLEX"   // queue 440
	LolGameModeAram   LolGameMode = "ARAM"   // map 12 (Howling Abyss)
	LolGameModeNormal LolGameMode = "NORMAL" // còn lại
)

type LolPosition string

const (
	LolPositionTop LolPosition = "TOP" // Riot teamPosition TOP
	LolPositionJgl LolPosition = "JGL" // JUNGLE
	LolPositionMid LolPosition = "MID" // MIDDLE
	LolPositionAdc LolPosition = "ADC" // BOTTOM
	LolPositionSpt LolPosition = "SPT" // UTILITY
	LolPositionUnk LolPosition = "UNK" // rỗng / giá trị lạ (ARAM, Arena...)
)

const (
	queueRankedSolo = 420
	queueRankedFlex = 440
	mapHowlingAbyss = 12
)

// Tạm thời, sau thay bằng công thức riêng.
const defaultPerfScore = 100

// ---------- entity ----------

type LolMatch struct {
	Id                string                `json:"id"`
	Server            RiotServer            `json:"server"`
	Mode              LolGameMode           `json:"mode"`
	Patch             string                `json:"patch"`
	GameStartAt       time.Time             `json:"gameStartAt"`
	DurationSec       int32                 `json:"durationSec"`
	IsRemake          bool                  `json:"isRemake"`
	EstimatedRank     shared.Nullable[Rank] `json:"estimatedRank"`
	BannedChampionIds []int32               `json:"bannedChampionIds"`
	WinningTeam       int16                 `json:"winningTeam"`
	Team1Kills        int16                 `json:"team1Kills"`
	Team1DragonKills  int16                 `json:"team1DragonKills"`
	Team1HeraldKills  int16                 `json:"team1HeraldKills"`
	Team1BaronKills   int16                 `json:"team1BaronKills"`
	Team2Kills        int16                 `json:"team2Kills"`
	Team2DragonKills  int16                 `json:"team2DragonKills"`
	Team2HeraldKills  int16                 `json:"team2HeraldKills"`
	Team2BaronKills   int16                 `json:"team2BaronKills"`
	CreatedAt         time.Time             `json:"createdAt"`
	Participants      []LolMatchParticipant `json:"participants"`
}

func ToLolMatch(m db.LolMatch, participants []LolMatchParticipant) LolMatch {
	return LolMatch{
		Id:                m.ID,
		Server:            RiotServer(m.Server),
		Mode:              LolGameMode(m.Mode),
		Patch:             m.Patch,
		GameStartAt:       m.GameStartAt.Time,
		DurationSec:       m.DurationSec,
		IsRemake:          m.IsRemake,
		EstimatedRank:     shared.NullableText[Rank](m.EstimatedRank),
		BannedChampionIds: m.BannedChampionIds,
		WinningTeam:       m.WinningTeam,
		Team1Kills:        m.Team1Kills,
		Team1DragonKills:  m.Team1DragonKills,
		Team1HeraldKills:  m.Team1HeraldKills,
		Team1BaronKills:   m.Team1BaronKills,
		Team2Kills:        m.Team2Kills,
		Team2DragonKills:  m.Team2DragonKills,
		Team2HeraldKills:  m.Team2HeraldKills,
		Team2BaronKills:   m.Team2BaronKills,
		CreatedAt:         m.CreatedAt.Time,
		Participants:      participants,
	}
}

type LolMatchParticipant struct {
	Team                 int16                  `json:"team"`
	IsWin                bool                   `json:"isWin"`
	PlayerId             string                 `json:"playerId"`
	RiotName             string                 `json:"riotName"`
	RiotTag              string                 `json:"riotTag"`
	RankPower            shared.Nullable[int32] `json:"rankPower"`
	ChampionId           int32                  `json:"championId"`
	ChampLevel           int16                  `json:"champLevel"`
	Position             LolPosition            `json:"position"`
	Kills                int16                  `json:"kills"`
	Deaths               int16                  `json:"deaths"`
	Assists              int16                  `json:"assists"`
	Kda                  float32                `json:"kda"`
	KillParticipation    float32                `json:"killParticipation"`
	GoldEarned           int32                  `json:"goldEarned"`
	MinionsKilled        int32                  `json:"minionsKilled"`
	NeutralMinionsKilled int32                  `json:"neutralMinionsKilled"`
	Cs                   int32                  `json:"cs"`
	DmgToChamps          int32                  `json:"dmgToChamps"`
	PhysicalDmgToChamps  int32                  `json:"physicalDmgToChamps"`
	MagicDmgToChamps     int32                  `json:"magicDmgToChamps"`
	TrueDmgToChamps      int32                  `json:"trueDmgToChamps"`
	DmgTaken             int32                  `json:"dmgTaken"`
	VisionScore          int32                  `json:"visionScore"`
	PerfScore            int32                  `json:"perfScore"`
	Spell1Id             int16                  `json:"spell1Id"`
	Spell2Id             int16                  `json:"spell2Id"`
	RunePrimaryStyle     int32                  `json:"runePrimaryStyle"`
	RuneSubStyle         int32                  `json:"runeSubStyle"`
	KeyRune              int32                  `json:"keyRune"`
	Runes                []int32                `json:"runes"`
	StatRunes            []int32                `json:"statRunes"`
	Items                []int32                `json:"items"`
}

func ToLolMatchParticipant(p db.LolMatchParticipant) LolMatchParticipant {
	return LolMatchParticipant{
		Team:                 p.Team,
		IsWin:                p.IsWin,
		PlayerId:             p.PlayerID,
		RiotName:             p.RiotName,
		RiotTag:              p.RiotTag,
		RankPower:            shared.NullableInt4(p.RankPower),
		ChampionId:           p.ChampionID,
		ChampLevel:           p.ChampLevel,
		Position:             LolPosition(p.Position),
		Kills:                p.Kills,
		Deaths:               p.Deaths,
		Assists:              p.Assists,
		Kda:                  p.Kda,
		KillParticipation:    p.KillParticipation,
		GoldEarned:           p.GoldEarned,
		MinionsKilled:        p.MinionsKilled,
		NeutralMinionsKilled: p.NeutralMinionsKilled,
		Cs:                   p.Cs,
		DmgToChamps:          p.DmgToChamps,
		PhysicalDmgToChamps:  p.PhysicalDmgToChamps,
		MagicDmgToChamps:     p.MagicDmgToChamps,
		TrueDmgToChamps:      p.TrueDmgToChamps,
		DmgTaken:             p.DmgTaken,
		VisionScore:          p.VisionScore,
		PerfScore:            p.PerfScore,
		Spell1Id:             p.Spell1ID,
		Spell2Id:             p.Spell2ID,
		RunePrimaryStyle:     p.RunePrimaryStyle,
		RuneSubStyle:         p.RuneSubStyle,
		KeyRune:              p.KeyRune,
		Runes:                p.Runes,
		StatRunes:            p.StatRunes,
		Items:                p.Items,
	}
}

// ---------- logic ----------

// Map MatchDto của Riot rồi lưu: 1 query lấy rank của mọi participant (rank_power / estimated_rank
// lấy từ solo rank của các participant đã có trong lol_players), insert tất cả match + participants
// bằng pgx batch trong 1 transaction, rồi đọc lại 1 lần. Match đã có thì không ghi đè (DO NOTHING).
// 1 match lỗi => không match nào được lưu.
// Kết quả theo game_start_at giảm dần.
func (a *Application) SaveLolMatchesToDb(ctx context.Context, matches []*external.MatchDto) ([]LolMatch, error) {
	if len(matches) == 0 {
		return []LolMatch{}, nil
	}

	puuids := []string{}
	for _, m := range matches {
		puuids = append(puuids, puuidsOf(m)...)
	}
	rankPowers, err := a.rankPowersOf(ctx, puuids)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(matches))
	matchParams := make([]db.InsertMatchesParams, 0, len(matches))
	participantParams := []db.InsertMatchParticipantsParams{}
	for _, m := range matches {
		mp, pps := insertParamsOf(m, rankPowers)
		ids = append(ids, mp.ID)
		matchParams = append(matchParams, mp)
		participantParams = append(participantParams, pps...)
	}

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)

	err = execBatch(q.InsertMatches(ctx, matchParams).Exec)
	if err != nil {
		return nil, err
	}
	err = execBatch(q.InsertMatchParticipants(ctx, participantParams).Exec)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return a.getLolMatchesByIds(ctx, ids)
}

// Lấy list match id của player từ Riot (start = offset, count = số match; 0 => Riot default 20, tối đa 100);
// match đã có trong DB thì đọc DB,
// còn lại gọi Riot (GetMatchesByIds, song song theo batch) rồi lưu tất cả 1 lần bằng SaveLolMatchesToDb.
// mode: NULL = mọi mode; chỉ nhận SOLO | FLEX (lọc theo queue của Riot), mode khác => err.
// Kết quả theo thứ tự Riot trả về (mới nhất trước).
// Không tìm thấy player => IsNull = true, err = nil.
func (a *Application) GetLolMatchesByPlayerInfo(ctx context.Context, server RiotServer, name, tag string, mode shared.Nullable[LolGameMode], start, count int) (shared.Nullable[[]LolMatch], error) {
	regions, ok := serverRegions[server]
	if !ok {
		return shared.Nullable[[]LolMatch]{}, fmt.Errorf("invalid server: %q", server)
	}
	opts := external.MatchIdsOptions{Start: start, Count: count}
	if !mode.IsNull {
		queue, ok := riotQueueOf(mode.Value)
		if !ok {
			return shared.Nullable[[]LolMatch]{}, fmt.Errorf("unsupported mode: %q", mode.Value)
		}
		opts.Queue = &queue
	}

	player, err := a.FindLolPlayerByPlayerInfo(ctx, server, name, tag)
	if err != nil {
		return shared.Nullable[[]LolMatch]{}, err
	}
	if player.IsNull {
		return shared.Nullable[[]LolMatch]{IsNull: true}, nil
	}

	matchIds, err := a.riot.GetMatchIdsByPuuid(ctx, regions.match, player.Value.Id, opts)
	if err != nil {
		return shared.Nullable[[]LolMatch]{}, err
	}

	existing, err := a.getLolMatchesByIds(ctx, matchIds)
	if err != nil {
		return shared.Nullable[[]LolMatch]{}, err
	}
	byId := map[string]LolMatch{}
	for _, m := range existing {
		byId[m.Id] = m
	}

	missingIds := []string{}
	for _, id := range matchIds {
		if _, ok := byId[id]; !ok {
			missingIds = append(missingIds, id)
		}
	}
	// Đợt 1: chỉ gọi Riot.
	dtos, err := a.riot.GetMatchesByIdsInParallel(ctx, regions.match, missingIds)
	if err != nil {
		return shared.Nullable[[]LolMatch]{}, err
	}

	// Đợt 2: lưu tất cả trong 1 lần.
	saved, err := a.SaveLolMatchesToDb(ctx, dtos)
	if err != nil {
		return shared.Nullable[[]LolMatch]{}, err
	}
	for _, m := range saved {
		byId[m.Id] = m
	}

	out := make([]LolMatch, 0, len(matchIds))
	for _, id := range matchIds {
		if m, ok := byId[id]; ok {
			out = append(out, m)
		}
	}
	return shared.Nullable[[]LolMatch]{Value: out}, nil
}

// ---------- private ----------

func lolGameModeOf(queueId, mapId int) LolGameMode {
	switch {
	case queueId == queueRankedSolo:
		return LolGameModeSolo
	case queueId == queueRankedFlex:
		return LolGameModeFlex
	case mapId == mapHowlingAbyss:
		return LolGameModeAram
	}
	return LolGameModeNormal
}

// Riot teamPosition => LolPosition. Rỗng / giá trị lạ (ARAM, Arena...) => UNK.
func lolPositionOf(teamPosition string) LolPosition {
	switch teamPosition {
	case "TOP":
		return LolPositionTop
	case "JUNGLE":
		return LolPositionJgl
	case "MIDDLE":
		return LolPositionMid
	case "BOTTOM":
		return LolPositionAdc
	case "UTILITY":
		return LolPositionSpt
	}
	return LolPositionUnk
}

// Queue id Riot để lọc match theo mode. ok = false nếu mode không lọc được bằng 1 queue (ARAM, NORMAL).
func riotQueueOf(mode LolGameMode) (queue int, ok bool) {
	switch mode {
	case LolGameModeSolo:
		return queueRankedSolo, true
	case LolGameModeFlex:
		return queueRankedFlex, true
	}
	return 0, false
}

// Riot teamId 100 => 1 (blue), 200 => 2 (red).
func teamOf(riotTeamId int) int16 {
	return int16(riotTeamId / 100)
}

// Trung bình rank power (đã bỏ unknown / UNRANKED) => rank ứng với bậc đó.
// Không có ai => NULL. Master trở lên LP không giới hạn nên bậc vượt Challenger thì lấy Challenger.
func estimatedRankOf(rankPowers []int32) shared.Nullable[Rank] {
	if len(rankPowers) == 0 {
		return shared.Nullable[Rank]{IsNull: true}
	}
	var sum int64
	for _, p := range rankPowers {
		sum += int64(p)
	}
	level := int32(sum / int64(len(rankPowers)) / rankPowerPerRank)
	level = max(rankLevels[RankIron], min(level, rankLevels[RankChallenger]))
	for rank, l := range rankLevels {
		if l == level {
			return shared.Nullable[Rank]{Value: rank}
		}
	}
	return shared.Nullable[Rank]{IsNull: true}
}

func puuidsOf(match *external.MatchDto) []string {
	out := make([]string, 0, len(match.Info.Participants))
	for _, p := range match.Info.Participants {
		out = append(out, p.Puuid)
	}
	return out
}

// puuid => solo rank power của player đã có trong lol_players. Không có trong DB => không có key.
func (a *Application) rankPowersOf(ctx context.Context, puuids []string) (map[string]shared.Nullable[int32], error) {
	players, err := a.q.GetPlayersByIds(ctx, puuids)
	if err != nil {
		return nil, err
	}
	out := map[string]shared.Nullable[int32]{}
	for _, p := range players {
		out[p.ID] = shared.NullableInt4(p.SoloRankPower)
	}
	return out, nil
}

// Map MatchDto thành params insert match + participants. rankPowers từ rankPowersOf.
func insertParamsOf(match *external.MatchDto, rankPowers map[string]shared.Nullable[int32]) (db.InsertMatchesParams, []db.InsertMatchParticipantsParams) {
	info := match.Info
	matchParams := db.InsertMatchesParams{
		ID:                match.Metadata.MatchId,
		Server:            info.PlatformId,
		Mode:              string(lolGameModeOf(info.QueueId, info.MapId)),
		Patch:             shared.PatchOf(info.GameVersion),
		GameStartAt:       pgtype.Timestamptz{Time: time.UnixMilli(info.GameStartTimestamp), Valid: true},
		DurationSec:       int32(info.GameDuration),
		BannedChampionIds: []int32{},
	}
	teamKills := map[int]int{} // Riot teamId => kills, dùng tính kill participation
	for _, t := range info.Teams {
		teamKills[t.TeamId] = t.Objectives.Champion.Kills
		for _, b := range t.Bans {
			if b.ChampionId > 0 {
				matchParams.BannedChampionIds = append(matchParams.BannedChampionIds, int32(b.ChampionId))
			}
		}
		team := teamOf(t.TeamId)
		if t.Win {
			matchParams.WinningTeam = team
		}
		switch team {
		case 1:
			matchParams.Team1Kills = int16(t.Objectives.Champion.Kills)
			matchParams.Team1DragonKills = int16(t.Objectives.Dragon.Kills)
			matchParams.Team1HeraldKills = int16(t.Objectives.RiftHerald.Kills)
			matchParams.Team1BaronKills = int16(t.Objectives.Baron.Kills)
		case 2:
			matchParams.Team2Kills = int16(t.Objectives.Champion.Kills)
			matchParams.Team2DragonKills = int16(t.Objectives.Dragon.Kills)
			matchParams.Team2HeraldKills = int16(t.Objectives.RiftHerald.Kills)
			matchParams.Team2BaronKills = int16(t.Objectives.Baron.Kills)
		}
	}

	participantParams := make([]db.InsertMatchParticipantsParams, 0, len(info.Participants))
	knownRankPowers := []int32{}
	for _, p := range info.Participants {
		if p.GameEndedInEarlySurrender {
			matchParams.IsRemake = true
		}
		rankPower, ok := rankPowers[p.Puuid]
		if !ok {
			rankPower = shared.Nullable[int32]{IsNull: true}
		}
		if !rankPower.IsNull && rankPower.Value > 0 { // bỏ unknown và UNRANKED
			knownRankPowers = append(knownRankPowers, rankPower.Value)
		}
		participantParams = append(participantParams, participantParamsOf(matchParams.ID, p, rankPower, teamKills[p.TeamId]))
	}
	matchParams.EstimatedRank = shared.PgText(estimatedRankOf(knownRankPowers))
	return matchParams, participantParams
}

// Chạy Exec của 1 batch sqlc (:batchexec), trả lỗi đầu tiên.
func execBatch(exec func(func(int, error))) error {
	var firstErr error
	exec(func(_ int, err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	})
	return firstErr
}

func participantParamsOf(matchId string, p external.ParticipantDto, rankPower shared.Nullable[int32], teamKills int) db.InsertMatchParticipantsParams {
	var primary, sub external.PerkStyleDto
	for _, s := range p.Perks.Styles {
		switch s.Description {
		case "primaryStyle":
			primary = s
		case "subStyle":
			sub = s
		}
	}
	runes := []int32{}
	for _, s := range primary.Selections {
		runes = append(runes, int32(s.Perk))
	}
	for _, s := range sub.Selections {
		runes = append(runes, int32(s.Perk))
	}
	var keyRune int32
	if len(runes) > 0 {
		keyRune = runes[0]
	}

	var killParticipation float32
	if teamKills > 0 {
		killParticipation = float32(p.Kills+p.Assists) / float32(teamKills)
	}

	return db.InsertMatchParticipantsParams{
		MatchID:              matchId,
		Team:                 teamOf(p.TeamId),
		IsWin:                p.Win,
		PlayerID:             p.Puuid,
		RiotName:             p.RiotIdGameName,
		RiotTag:              p.RiotIdTagline,
		RankPower:            shared.PgInt4(rankPower),
		ChampionID:           int32(p.ChampionId),
		ChampLevel:           int16(p.ChampLevel),
		Position:             string(lolPositionOf(p.TeamPosition)),
		Kills:                int16(p.Kills),
		Deaths:               int16(p.Deaths),
		Assists:              int16(p.Assists),
		Kda:                  float32(p.Kills+p.Assists) / float32(max(p.Deaths, 1)),
		KillParticipation:    killParticipation,
		GoldEarned:           int32(p.GoldEarned),
		MinionsKilled:        int32(p.TotalMinionsKilled),
		NeutralMinionsKilled: int32(p.NeutralMinionsKilled),
		Cs:                   int32(p.TotalMinionsKilled + p.NeutralMinionsKilled),
		DmgToChamps:          int32(p.TotalDamageDealtToChampions),
		PhysicalDmgToChamps:  int32(p.PhysicalDamageDealtToChampions),
		MagicDmgToChamps:     int32(p.MagicDamageDealtToChampions),
		TrueDmgToChamps:      int32(p.TrueDamageDealtToChampions),
		DmgTaken:             int32(p.TotalDamageTaken),
		VisionScore:          int32(p.VisionScore),
		PerfScore:            defaultPerfScore,
		// sort để Flash+Ignite == Ignite+Flash khi GROUP BY
		Spell1ID:         int16(min(p.Summoner1Id, p.Summoner2Id)),
		Spell2ID:         int16(max(p.Summoner1Id, p.Summoner2Id)),
		RunePrimaryStyle: int32(primary.Style),
		RuneSubStyle:     int32(sub.Style),
		KeyRune:          keyRune,
		Runes:            runes,
		StatRunes:        []int32{int32(p.Perks.StatPerks.Offense), int32(p.Perks.StatPerks.Flex), int32(p.Perks.StatPerks.Defense)},
		Items:            []int32{int32(p.Item0), int32(p.Item1), int32(p.Item2), int32(p.Item3), int32(p.Item4), int32(p.Item5), int32(p.Item6)},
	}
}

// Đọc matches rồi participants của chúng (2 query), ghép theo match id.
// Id không có trong DB thì bỏ qua. Thứ tự theo game_start_at giảm dần.
func (a *Application) getLolMatchesByIds(ctx context.Context, ids []string) ([]LolMatch, error) {
	matchRows, err := a.q.GetMatchesByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	participantRows, err := a.q.GetMatchParticipantsByMatchIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	participantsByMatchId := map[string][]LolMatchParticipant{}
	for _, p := range participantRows {
		participantsByMatchId[p.MatchID] = append(participantsByMatchId[p.MatchID], ToLolMatchParticipant(p))
	}
	out := make([]LolMatch, 0, len(matchRows))
	for _, m := range matchRows {
		out = append(out, ToLolMatch(m, participantsByMatchId[m.ID]))
	}
	return out, nil
}
