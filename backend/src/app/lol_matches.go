package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/src/db"
	"backend/src/external"
	"backend/src/shared"

	"github.com/jackc/pgx/v5/pgtype"
)

// ---------- entity ----------

type Match struct {
	Id                string                `json:"id"`
	Server            Server                `json:"server"`
	Mode              GameMode              `json:"mode"`
	Patch             string                `json:"patch"`
	GameStartAt       time.Time             `json:"gameStartAt"`
	DurationSec       int32                 `json:"durationSec"`
	IsRemake          bool                  `json:"isRemake"`
	EstimatedRank     Rank                  `json:"estimatedRank"`
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
	Participants      []MatchParticipant    `json:"participants"`
}

func ToMatch(m db.LolMatch, participants []MatchParticipant) Match {
	return Match{
		Id:                m.ID,
		Server:            Server(m.Server),
		Mode:              GameMode(m.Mode),
		Patch:             m.Patch,
		GameStartAt:       m.GameStartAt.Time,
		DurationSec:       m.DurationSec,
		IsRemake:          m.IsRemake,
		EstimatedRank:     Rank(m.EstimatedRank),
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

type MatchParticipant struct {
	MatchId              string                 `json:"matchId"`
	Team                 int16                  `json:"team"`
	IsWin                bool                   `json:"isWin"`
	PlayerId             string                 `json:"playerId"`
	Name                 string                 `json:"name"`
	Tag                  string                 `json:"tag"`
	RankPower            shared.Nullable[int32] `json:"rankPower"`
	ChampionId           int32                  `json:"championId"`
	ChampionSlug         string                 `json:"championSlug"`
	ChampLevel           int16                  `json:"champLevel"`
	Position             Position               `json:"position"`
	Kills                int16                  `json:"kills"`
	Deaths               int16                  `json:"deaths"`
	Assists              int16                  `json:"assists"`
	Kda                  float32                `json:"kda"`
	KillParticipation    float32                `json:"killParticipation"`
	DoubleKills          int16                  `json:"doubleKills"`
	TripleKills          int16                  `json:"tripleKills"`
	QuadraKills          int16                  `json:"quadraKills"`
	PentaKills           int16                  `json:"pentaKills"`
	SoloKills            int16                  `json:"soloKills"`
	Gold                 int32                  `json:"gold"`
	GoldPerMin           float32                `json:"goldPerMin"`
	MinionsKilled        int32                  `json:"minionsKilled"`
	NeutralMinionsKilled int32                  `json:"neutralMinionsKilled"`
	Cs                   int32                  `json:"cs"`
	CsPerMin             float32                `json:"csPerMin"`
	DmgDealt             int32                  `json:"dmgDealt"`
	DmgPerMin            float32                `json:"dmgPerMin"`
	PhysicalDmgDealt     int32                  `json:"physicalDmgDealt"`
	MagicDmgDealt        int32                  `json:"magicDmgDealt"`
	TrueDmgDealt         int32                  `json:"trueDmgDealt"`
	DmgToTurrets         int32                  `json:"dmgToTurrets"`
	DmgTaken             int32                  `json:"dmgTaken"`
	Heal                 int32                  `json:"heal"`
	HealOthers           int32                  `json:"healOthers"`
	ShieldOthers         int32                  `json:"shieldOthers"`
	VisionScore          int32                  `json:"visionScore"`
	WardsPlaced          int32                  `json:"wardsPlaced"`
	WardsKilled          int32                  `json:"wardsKilled"`
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

func ToMatchParticipant(p db.LolMatchParticipant) MatchParticipant {
	return MatchParticipant{
		MatchId:              p.MatchID,
		Team:                 p.Team,
		IsWin:                p.IsWin,
		PlayerId:             p.PlayerID,
		Name:                 p.Name,
		Tag:                  p.Tag,
		RankPower:            shared.NullableInt4(p.RankPower),
		ChampionId:           p.ChampionID,
		ChampionSlug:         p.ChampionSlug,
		ChampLevel:           p.ChampLevel,
		Position:             Position(p.Position),
		Kills:                p.Kills,
		Deaths:               p.Deaths,
		Assists:              p.Assists,
		Kda:                  p.Kda,
		KillParticipation:    p.KillParticipation,
		DoubleKills:          p.DoubleKills,
		TripleKills:          p.TripleKills,
		QuadraKills:          p.QuadraKills,
		PentaKills:           p.PentaKills,
		SoloKills:            p.SoloKills,
		Gold:                 p.Gold,
		GoldPerMin:           p.GoldPerMin,
		MinionsKilled:        p.MinionsKilled,
		NeutralMinionsKilled: p.NeutralMinionsKilled,
		Cs:                   p.Cs,
		CsPerMin:             p.CsPerMin,
		DmgDealt:             p.DmgDealt,
		DmgPerMin:            p.DmgPerMin,
		PhysicalDmgDealt:     p.PhysicalDmgDealt,
		MagicDmgDealt:        p.MagicDmgDealt,
		TrueDmgDealt:         p.TrueDmgDealt,
		DmgToTurrets:         p.DmgToTurrets,
		DmgTaken:             p.DmgTaken,
		Heal:                 p.Heal,
		HealOthers:           p.HealOthers,
		ShieldOthers:         p.ShieldOthers,
		VisionScore:          p.VisionScore,
		WardsPlaced:          p.WardsPlaced,
		WardsKilled:          p.WardsKilled,
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
func (a *Application) SaveMatchesToDb(ctx context.Context, matches []*external.MatchDto) ([]Match, error) {
	if len(matches) == 0 {
		return []Match{}, nil
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

	return a.getMatchesByIds(ctx, ids)
}

// Phân giải player (DB / Riot) rồi giao phần còn lại cho GetMatchesByPuuid.
// mode: NULL = mọi mode; chỉ nhận SOLO | FLEX, mode khác => err.
// Không tìm thấy player => IsNull = true, err = nil.
func (a *Application) GetMatchesByPlayerInfo(ctx context.Context, server Server, name, tag string, mode shared.Nullable[GameMode], start, count int) (shared.Nullable[[]Match], error) {
	player, err := a.FindPlayerByPlayerInfo(ctx, server, name, tag)
	if err != nil {
		return shared.Nullable[[]Match]{}, err
	}
	if player.IsNull {
		return shared.Nullable[[]Match]{IsNull: true}, nil
	}

	matches, err := a.GetMatchesByPuuid(ctx, server, player.Value.Id, mode, start, count)
	if err != nil {
		return shared.Nullable[[]Match]{}, err
	}
	return shared.Nullable[[]Match]{Value: matches}, nil
}

// Lấy list match id của player từ Riot (start = offset, count = số match; 0 => Riot default 20, tối đa 100);
// match đã có trong DB thì đọc DB,
// còn lại gọi Riot (GetMatchesByIds, song song theo batch) rồi lưu tất cả 1 lần bằng SaveMatchesToDb.
// mode: NULL = mọi mode; chỉ nhận SOLO | FLEX (lọc theo queue của Riot), mode khác => err.
// Kết quả theo thứ tự Riot trả về (mới nhất trước).
func (a *Application) GetMatchesByPuuid(ctx context.Context, server Server, puuid string, mode shared.Nullable[GameMode], start, count int) ([]Match, error) {
	if !server.IsValid() {
		return nil, fmt.Errorf("invalid server: %q", server)
	}
	routing := server.riot()
	opts := external.MatchIdsOptions{Start: start, Count: count}
	if !mode.IsNull {
		queue, ok := riotQueueOf(mode.Value)
		if !ok {
			return nil, fmt.Errorf("unsupported mode: %q", mode.Value)
		}
		opts.Queue = &queue
	}

	matchIds, err := a.riot.GetMatchIdsByPuuid(ctx, routing.matchRegion, puuid, opts)
	if err != nil {
		return nil, err
	}

	existing, err := a.getMatchesByIds(ctx, matchIds)
	if err != nil {
		return nil, err
	}
	byId := map[string]Match{}
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
	dtos, err := a.riot.GetMatchesByIdsInParallel(ctx, routing.matchRegion, missingIds)
	if err != nil {
		return nil, err
	}

	// Đợt 2: lưu tất cả trong 1 lần.
	saved, err := a.SaveMatchesToDb(ctx, dtos)
	if err != nil {
		return nil, err
	}
	for _, m := range saved {
		byId[m.Id] = m
	}

	out := make([]Match, 0, len(matchIds))
	for _, id := range matchIds {
		if m, ok := byId[id]; ok {
			out = append(out, m)
		}
	}
	return out, nil
}

// ---------- private ----------

// Riot teamId 100 => 1 (blue), 200 => 2 (red).
func teamOf(riotTeamId int) int16 {
	return int16(riotTeamId / 100)
}

// Giá trị trên mỗi phút, để so được giữa các trận dài ngắn khác nhau.
// Duration <= 0 (dữ liệu lỗi) => 0.
func perMinuteOf(value int, durationSec int32) float32 {
	if durationSec <= 0 {
		return 0
	}
	return float32(value) * 60 / float32(durationSec)
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
	// Riot trả platform id ("VN2"), DB lưu enum Server của app ("VN"). Platform lạ (Riot mở server
	// mới) thì giữ nguyên giá trị Riot để không mất dữ liệu, dù nó không khớp enum nào.
	server := strings.ToUpper(info.PlatformId)
	if s := serverOf(info.PlatformId); !s.IsNull {
		server = string(s.Value)
	}
	matchParams := db.InsertMatchesParams{
		ID:                match.Metadata.MatchId,
		Server:            server,
		Mode:              string(gameModeOf(info.QueueId, info.MapId)),
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
		participantParams = append(participantParams, participantParamsOf(matchParams.ID, p, rankPower, teamKills[p.TeamId], matchParams.DurationSec, laneOpponentGoldOf(p, info.Participants)))
	}
	matchParams.EstimatedRank = string(estimatedRankOf(knownRankPowers))
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

func participantParamsOf(matchId string, p external.ParticipantDto, rankPower shared.Nullable[int32], teamKills int, durationSec int32, laneOpponentGold shared.Nullable[int]) db.InsertMatchParticipantsParams {
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

	position := positionOf(p.TeamPosition)
	// Dùng killParticipation đã ép float32 để tính lại từ DB (cột REAL) ra đúng điểm này.
	perfScore := perfScoreOf(perfScoreInput{
		Position:          position,
		KillParticipation: float64(killParticipation),
		Deaths:            p.Deaths,
		IsWin:             p.Win,
		GoldEarned:        p.GoldEarned,
		LaneOpponentGold:  laneOpponentGold,
		DurationSec:       durationSec,
	})

	cs := p.TotalMinionsKilled + p.NeutralMinionsKilled

	return db.InsertMatchParticipantsParams{
		MatchID:              matchId,
		Team:                 teamOf(p.TeamId),
		IsWin:                p.Win,
		PlayerID:             p.Puuid,
		Name:                 p.RiotIdGameName,
		Tag:                  p.RiotIdTagline,
		RankPower:            shared.PgInt4(rankPower),
		ChampionID:           int32(p.ChampionId),
		ChampionSlug:         p.ChampionName,
		ChampLevel:           int16(p.ChampLevel),
		Position:             string(position),
		Kills:                int16(p.Kills),
		Deaths:               int16(p.Deaths),
		Assists:              int16(p.Assists),
		Kda:                  float32(p.Kills+p.Assists) / float32(max(p.Deaths, 1)),
		KillParticipation:    killParticipation,
		DoubleKills:          int16(p.DoubleKills),
		TripleKills:          int16(p.TripleKills),
		QuadraKills:          int16(p.QuadraKills),
		PentaKills:           int16(p.PentaKills),
		SoloKills:            int16(p.Challenges.SoloKills),
		Gold:                 int32(p.GoldEarned),
		GoldPerMin:           perMinuteOf(p.GoldEarned, durationSec),
		MinionsKilled:        int32(p.TotalMinionsKilled),
		NeutralMinionsKilled: int32(p.NeutralMinionsKilled),
		Cs:                   int32(cs),
		CsPerMin:             perMinuteOf(cs, durationSec),
		DmgDealt:             int32(p.TotalDamageDealtToChampions),
		DmgPerMin:            perMinuteOf(p.TotalDamageDealtToChampions, durationSec),
		PhysicalDmgDealt:     int32(p.PhysicalDamageDealtToChampions),
		MagicDmgDealt:        int32(p.MagicDamageDealtToChampions),
		TrueDmgDealt:         int32(p.TrueDamageDealtToChampions),
		DmgToTurrets:         int32(p.DamageDealtToTurrets),
		DmgTaken:             int32(p.TotalDamageTaken),
		Heal:                 int32(p.TotalHeal),
		HealOthers:           int32(p.TotalHealsOnTeammates),
		ShieldOthers:         int32(p.TotalDamageShieldedOnTeammates),
		VisionScore:          int32(p.VisionScore),
		WardsPlaced:          int32(p.WardsPlaced),
		WardsKilled:          int32(p.WardsKilled),
		PerfScore:            perfScore,
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
func (a *Application) getMatchesByIds(ctx context.Context, ids []string) ([]Match, error) {
	matchRows, err := a.q.GetMatchesByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	participantRows, err := a.q.GetMatchParticipantsByMatchIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	participantsByMatchId := map[string][]MatchParticipant{}
	for _, p := range participantRows {
		participantsByMatchId[p.MatchID] = append(participantsByMatchId[p.MatchID], ToMatchParticipant(p))
	}
	out := make([]Match, 0, len(matchRows))
	for _, m := range matchRows {
		out = append(out, ToMatch(m, participantsByMatchId[m.ID]))
	}
	return out, nil
}
