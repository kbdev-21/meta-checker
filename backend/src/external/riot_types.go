package external

// ---------- params ----------

type QueueType string

const (
	QueueRankedSolo   QueueType = "RANKED_SOLO_5x5"
	QueueRankedFlexSR QueueType = "RANKED_FLEX_SR"
	QueueRankedFlexTT QueueType = "RANKED_FLEX_TT"
)

// Zero value = không gửi param đó.
type MatchIdsOptions struct {
	StartTime int64  // epoch seconds
	EndTime   int64  // epoch seconds
	Queue     *int   // queue id, vd 420 = ranked solo
	Type      string // ranked | normal | tourney | tutorial
	Start     int    // default 0
	Count     int    // default 20, 0-100
}

// ---------- account-v1 ----------

type AccountDto struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// ---------- league-v4 ----------

type MiniSeriesDto struct {
	Losses   int    `json:"losses"`
	Progress string `json:"progress"`
	Target   int    `json:"target"`
	Wins     int    `json:"wins"`
}

type LeagueEntryDto struct {
	LeagueId     string         `json:"leagueId"`
	Puuid        string         `json:"puuid"`
	QueueType    string         `json:"queueType"`
	Tier         string         `json:"tier"`
	Rank         string         `json:"rank"`
	LeaguePoints int            `json:"leaguePoints"`
	Wins         int            `json:"wins"`
	Losses       int            `json:"losses"`
	HotStreak    bool           `json:"hotStreak"`
	Veteran      bool           `json:"veteran"`
	FreshBlood   bool           `json:"freshBlood"`
	Inactive     bool           `json:"inactive"`
	MiniSeries   *MiniSeriesDto `json:"miniSeries"`
}

type LeagueItemDto struct {
	Puuid        string         `json:"puuid"`
	Rank         string         `json:"rank"`
	LeaguePoints int            `json:"leaguePoints"`
	Wins         int            `json:"wins"`
	Losses       int            `json:"losses"`
	HotStreak    bool           `json:"hotStreak"`
	Veteran      bool           `json:"veteran"`
	FreshBlood   bool           `json:"freshBlood"`
	Inactive     bool           `json:"inactive"`
	MiniSeries   *MiniSeriesDto `json:"miniSeries"`
}

type LeagueListDto struct {
	LeagueId string          `json:"leagueId"`
	Tier     string          `json:"tier"`
	Name     string          `json:"name"`
	Queue    string          `json:"queue"`
	Entries  []LeagueItemDto `json:"entries"`
}

// ---------- summoner-v4 ----------

type SummonerDto struct {
	Puuid         string `json:"puuid"`
	ProfileIconId int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int64  `json:"summonerLevel"`
}

// ---------- match-v5: match ----------

type MatchDto struct {
	Metadata MatchMetadataDto `json:"metadata"`
	Info     MatchInfoDto     `json:"info"`
}

type MatchMetadataDto struct {
	DataVersion  string   `json:"dataVersion"`
	MatchId      string   `json:"matchId"`
	Participants []string `json:"participants"`
}

type MatchInfoDto struct {
	GameId             int64            `json:"gameId"`
	PlatformId         string           `json:"platformId"`
	QueueId            int              `json:"queueId"`
	MapId              int              `json:"mapId"`
	GameMode           string           `json:"gameMode"`
	GameType           string           `json:"gameType"`
	GameVersion        string           `json:"gameVersion"`
	GameCreation       int64            `json:"gameCreation"`
	GameStartTimestamp int64            `json:"gameStartTimestamp"`
	GameEndTimestamp   int64            `json:"gameEndTimestamp"`
	GameDuration       int64            `json:"gameDuration"` // seconds
	EndOfGameResult    string           `json:"endOfGameResult"`
	Participants       []ParticipantDto `json:"participants"`
	Teams              []TeamDto        `json:"teams"`
}

type ParticipantDto struct {
	Puuid          string `json:"puuid"`
	RiotIdGameName string `json:"riotIdGameName"`
	RiotIdTagline  string `json:"riotIdTagline"`
	ParticipantId  int    `json:"participantId"`
	TeamId         int    `json:"teamId"`
	Win            bool   `json:"win"`

	ChampionId         int    `json:"championId"`
	ChampionName       string `json:"championName"`
	ChampLevel         int    `json:"champLevel"`
	TeamPosition       string `json:"teamPosition"`
	IndividualPosition string `json:"individualPosition"`
	Lane               string `json:"lane"`
	Role               string `json:"role"`

	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`

	GoldEarned                     int `json:"goldEarned"`
	TotalMinionsKilled             int `json:"totalMinionsKilled"`
	NeutralMinionsKilled           int `json:"neutralMinionsKilled"`
	TotalDamageDealtToChampions    int `json:"totalDamageDealtToChampions"`
	PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`
	MagicDamageDealtToChampions    int `json:"magicDamageDealtToChampions"`
	TrueDamageDealtToChampions     int `json:"trueDamageDealtToChampions"`
	TotalDamageTaken               int `json:"totalDamageTaken"`
	VisionScore                    int `json:"visionScore"`

	GameEndedInEarlySurrender bool `json:"gameEndedInEarlySurrender"` // remake

	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`

	Summoner1Id int      `json:"summoner1Id"`
	Summoner2Id int      `json:"summoner2Id"`
	Perks       PerksDto `json:"perks"`
}

type PerksDto struct {
	StatPerks PerkStatsDto   `json:"statPerks"`
	Styles    []PerkStyleDto `json:"styles"`
}

type PerkStatsDto struct {
	Defense int `json:"defense"`
	Flex    int `json:"flex"`
	Offense int `json:"offense"`
}

type PerkStyleDto struct {
	Description string                  `json:"description"` // primaryStyle | subStyle
	Style       int                     `json:"style"`
	Selections  []PerkStyleSelectionDto `json:"selections"`
}

type PerkStyleSelectionDto struct {
	Perk int `json:"perk"`
	Var1 int `json:"var1"`
	Var2 int `json:"var2"`
	Var3 int `json:"var3"`
}

type TeamDto struct {
	TeamId     int           `json:"teamId"`
	Win        bool          `json:"win"`
	Bans       []BanDto      `json:"bans"`
	Objectives ObjectivesDto `json:"objectives"`
}

type BanDto struct {
	ChampionId int `json:"championId"`
	PickTurn   int `json:"pickTurn"`
}

type ObjectivesDto struct {
	Baron      ObjectiveDto `json:"baron"`
	Champion   ObjectiveDto `json:"champion"`
	Dragon     ObjectiveDto `json:"dragon"`
	Horde      ObjectiveDto `json:"horde"`
	Inhibitor  ObjectiveDto `json:"inhibitor"`
	RiftHerald ObjectiveDto `json:"riftHerald"`
	Tower      ObjectiveDto `json:"tower"`
	Atakhan    ObjectiveDto `json:"atakhan"`
}

type ObjectiveDto struct {
	First bool `json:"first"`
	Kills int  `json:"kills"`
}

// ---------- match-v5: timeline ----------

type TimelineDto struct {
	Metadata MatchMetadataDto `json:"metadata"`
	Info     TimelineInfoDto  `json:"info"`
}

type TimelineInfoDto struct {
	GameId        int64                    `json:"gameId"`
	FrameInterval int64                    `json:"frameInterval"`
	Participants  []TimelineParticipantDto `json:"participants"`
	Frames        []FrameDto               `json:"frames"`
}

type TimelineParticipantDto struct {
	ParticipantId int    `json:"participantId"`
	Puuid         string `json:"puuid"`
}

type FrameDto struct {
	Timestamp         int64                          `json:"timestamp"`
	Events            []EventDto                     `json:"events"`
	ParticipantFrames map[string]ParticipantFrameDto `json:"participantFrames"` // key "1".."10"
}

type ParticipantFrameDto struct {
	ParticipantId       int         `json:"participantId"`
	Level               int         `json:"level"`
	Xp                  int         `json:"xp"`
	CurrentGold         int         `json:"currentGold"`
	TotalGold           int         `json:"totalGold"`
	MinionsKilled       int         `json:"minionsKilled"`
	JungleMinionsKilled int         `json:"jungleMinionsKilled"`
	Position            PositionDto `json:"position"`
}

type PositionDto struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Các field optional tùy theo Type (ITEM_PURCHASED, SKILL_LEVEL_UP, CHAMPION_KILL, ELITE_MONSTER_KILL, BUILDING_KILL, ...).
type EventDto struct {
	Timestamp               int64        `json:"timestamp"`
	Type                    string       `json:"type"`
	ParticipantId           int          `json:"participantId"`
	ItemId                  int          `json:"itemId"`
	AfterId                 int          `json:"afterId"`
	BeforeId                int          `json:"beforeId"`
	SkillSlot               int          `json:"skillSlot"`
	LevelUpType             string       `json:"levelUpType"`
	Level                   int          `json:"level"`
	KillerId                int          `json:"killerId"`
	VictimId                int          `json:"victimId"`
	AssistingParticipantIds []int        `json:"assistingParticipantIds"`
	KillerTeamId            int          `json:"killerTeamId"`
	TeamId                  int          `json:"teamId"`
	MonsterType             string       `json:"monsterType"`
	MonsterSubType          string       `json:"monsterSubType"`
	BuildingType            string       `json:"buildingType"`
	TowerType               string       `json:"towerType"`
	LaneType                string       `json:"laneType"`
	WinningTeam             int          `json:"winningTeam"`
	Position                *PositionDto `json:"position"`
}
