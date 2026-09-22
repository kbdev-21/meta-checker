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

	DoubleKills int `json:"doubleKills"`
	TripleKills int `json:"tripleKills"`
	QuadraKills int `json:"quadraKills"`
	PentaKills  int `json:"pentaKills"`

	GoldEarned                     int `json:"goldEarned"`
	TotalMinionsKilled             int `json:"totalMinionsKilled"`
	NeutralMinionsKilled           int `json:"neutralMinionsKilled"`
	TotalDamageDealtToChampions    int `json:"totalDamageDealtToChampions"`
	PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`
	MagicDamageDealtToChampions    int `json:"magicDamageDealtToChampions"`
	TrueDamageDealtToChampions     int `json:"trueDamageDealtToChampions"`
	DamageDealtToTurrets           int `json:"damageDealtToTurrets"`
	TotalDamageTaken               int `json:"totalDamageTaken"`
	TotalHeal                      int `json:"totalHeal"`
	TotalHealsOnTeammates          int `json:"totalHealsOnTeammates"`
	TotalDamageShieldedOnTeammates int `json:"totalDamageShieldedOnTeammates"`
	VisionScore                    int `json:"visionScore"`
	WardsPlaced                    int `json:"wardsPlaced"`
	WardsKilled                    int `json:"wardsKilled"`

	GameEndedInEarlySurrender bool `json:"gameEndedInEarlySurrender"` // remake

	Item0 int `json:"item0"`
	Item1 int `json:"item1"`
	Item2 int `json:"item2"`
	Item3 int `json:"item3"`
	Item4 int `json:"item4"`
	Item5 int `json:"item5"`
	Item6 int `json:"item6"`

	Summoner1Id int           `json:"summoner1Id"`
	Summoner2Id int           `json:"summoner2Id"`
	Perks       PerksDto      `json:"perks"`
	Challenges  ChallengesDto `json:"challenges"`
}

type ChallengesDto struct {
	SoloKills int `json:"soloKills"`
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
