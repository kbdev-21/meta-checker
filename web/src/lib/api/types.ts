// Type của mọi payload backend trả về. Mirror của các struct trong backend/src/app.
// Quy ước: Nullable[T] của Go => T | null; time.Time => string (RFC3339).

// ---------- enum ----------

// Server của app, KHÔNG phải platform id của Riot ("VN" chứ không phải "VN2").
export const SERVERS = [
	'BR',
	'EUNE',
	'EUW',
	'JP',
	'KR',
	'LAN',
	'LAS',
	'ME',
	'NA',
	'OCE',
	'RU',
	'SG',
	'TR',
	'TW',
	'VN'
] as const;
export type Server = (typeof SERVERS)[number];

// Rank trong app = "tier" của Riot.
export const RANKS = [
	'UNRANKED',
	'IRON',
	'BRONZE',
	'SILVER',
	'GOLD',
	'PLATINUM',
	'EMERALD',
	'DIAMOND',
	'MASTER',
	'GRANDMASTER',
	'CHALLENGER'
] as const;
export type Rank = (typeof RANKS)[number];

// Tier trong app = "rank" của Riot.
export const TIERS = ['I', 'II', 'III', 'IV'] as const;
export type Tier = (typeof TIERS)[number];

export const GAME_MODES = ['SOLO', 'FLEX', 'ARAM', 'NORMAL'] as const;
export type GameMode = (typeof GAME_MODES)[number];

// Mode lọc được khi list match: backend chỉ nhận SOLO | FLEX (lọc theo queue của Riot).
export const MATCH_LIST_MODES = ['SOLO', 'FLEX'] as const;
export type MatchListMode = (typeof MATCH_LIST_MODES)[number];

export const POSITIONS = ['TOP', 'JGL', 'MID', 'ADC', 'SPT', 'UNK'] as const;
export type Position = (typeof POSITIONS)[number];

export const ITEM_TYPES = [
	'CONSUMABLE',
	'TRINKET',
	'BOOTS',
	'STARTER',
	'BASIC',
	'EPIC',
	'LEGENDARY'
] as const;
export type ItemType = (typeof ITEM_TYPES)[number];

export const RANK_BUCKETS = ['MASTER_PLUS'] as const;
export type RankBucket = (typeof RANK_BUCKETS)[number];

export const CHAMPION_TIERS = ['S', 'A', 'B', 'C', 'D'] as const;
export type ChampionTier = (typeof CHAMPION_TIERS)[number];

// server đặc biệt của meta: gộp mọi server.
export const META_SERVER_GLOBAL = 'GLOBAL';

// KHÁC với SERVERS: tra cứu player chạy được mọi server (backend gọi thẳng Riot), nhưng
// meta chỉ được tổng hợp cho những server có crawler (app.Crawlers ở backend) — xin server
// ngoài danh sách này sẽ luôn ra 404. Thêm server mới ở đây khi backend thêm crawler.
export const META_SERVERS = [META_SERVER_GLOBAL, 'VN', 'KR', 'EUW', 'EUNE', 'NA', 'BR'] as const;
export type MetaServer = (typeof META_SERVERS)[number];

// ---------- lol-data ----------

export type Champion = {
	id: number;
	slug: string;
	name: string;
	title: string;
	imgUrl: string;
	patch: string;
	updatedAt: string;
};

export type Item = {
	id: number;
	name: string;
	plaintext: string;
	type: ItemType;
	goldTotal: number;
	fromItems: number[];
	intoItems: number[];
	isSummonersRift: boolean;
	imgUrl: string;
	patch: string;
	updatedAt: string;
};

export type Spell = {
	id: number;
	slug: string;
	name: string;
	description: string;
	imgUrl: string;
	patch: string;
	updatedAt: string;
};

// Gộp cả cây rune lẫn rune. styleId === null => dòng này là một cây.
export type Rune = {
	id: number;
	styleId: number | null;
	slot: number | null;
	slug: string;
	name: string;
	shortDesc: string;
	imgUrl: string;
	patch: string;
	updatedAt: string;
};

// ---------- player ----------

export type Player = {
	id: string;
	server: Server;
	name: string;
	tag: string;
	profileIconId: number | null;
	level: number | null;
	soloRank: Rank | null;
	soloTier: Tier | null;
	soloLp: number;
	soloRankPower: number | null;
	soloWins: number;
	soloLosses: number;
	flexRank: Rank | null;
	flexTier: Tier | null;
	flexLp: number;
	flexWins: number;
	flexLosses: number;
	createdAt: string;
	updatedAt: string;
};

// ---------- match ----------

export type Match = {
	id: string;
	server: Server;
	mode: GameMode;
	patch: string;
	gameStartAt: string;
	durationSec: number;
	isRemake: boolean;
	estimatedRank: Rank;
	bannedChampionIds: number[];
	winningTeam: number;
	team1Kills: number;
	team1DragonKills: number;
	team1HeraldKills: number;
	team1BaronKills: number;
	team2Kills: number;
	team2DragonKills: number;
	team2HeraldKills: number;
	team2BaronKills: number;
	createdAt: string;
	participants: MatchParticipant[];
};

export type MatchParticipant = {
	matchId: string;
	team: number;
	isWin: boolean;
	playerId: string;
	participantId: number; // participantId 1..10 của Riot
	name: string;
	tag: string;
	rankPower: number | null;
	championId: number;
	championSlug: string;
	champLevel: number;
	position: Position;
	kills: number;
	deaths: number;
	assists: number;
	kda: number;
	killParticipation: number;
	doubleKills: number;
	tripleKills: number;
	quadraKills: number;
	pentaKills: number;
	soloKills: number;
	gold: number;
	goldPerMin: number;
	minionsKilled: number;
	neutralMinionsKilled: number;
	cs: number;
	csPerMin: number;
	dmgDealt: number;
	dmgPerMin: number;
	physicalDmgDealt: number;
	magicDmgDealt: number;
	trueDmgDealt: number;
	dmgToTurrets: number;
	dmgTaken: number;
	dmgTakenPerMin: number;
	crowdControl: number; // timeCCingOthers của Riot: điểm CC có trọng số (giống op.gg)
	ccPerMin: number;
	heal: number;
	healOthers: number;
	shieldOthers: number;
	visionScore: number;
	wardsPlaced: number;
	wardsKilled: number;
	perfScore: number;
	spell1Id: number;
	spell2Id: number;
	runePrimaryStyle: number;
	runeSubStyle: number;
	keyRune: number;
	runes: number[];
	statRunes: number[];
	items: number[];

	// Từ timeline. Riot không có timeline => mảng rỗng / 0.
	starterSets: number[]; // item mua trong 60s đầu, bỏ trinket, sort tăng dần
	skillsLeveled: number[]; // skillSlot theo thứ tự lên: 1 = Q, 2 = W, 3 = E, 4 = R
	firstLegendItem: number; // 0 = chưa xong đồ legendary nào
	legendItemsPurchased: number[]; // đồ legendary theo thứ tự mua
};

// ---------- analytics ----------

// Số đếm chung của mọi phần tử build, backend nhúng phẳng vào từng stat.
export type GameStat = {
	games: number;
	wins: number;
};

export type SpellComboStat = GameStat & {
	spell1Id: number;
	spell2Id: number;
};

export type RuneStat = GameStat & {
	runePrimaryStyle: number;
	runeSubStyle: number;
	keyRune: number;
	runes: number[];
	statRunes: number[];
};

export type ItemStat = GameStat & {
	itemId: number;
};

export type MatchupStat = GameStat & {
	opponentChampionId: number;
};

// Một bộ item: starter set (đã sort) hoặc 3 đồ legendary đầu (giữ thứ tự mua).
export type ItemSetStat = GameStat & {
	itemIds: number[];
};

// Thứ tự lên skill (13 lần đầu), mỗi phần tử là skillSlot: 1 = Q, 2 = W, 3 = E, 4 = R.
export type SkillOrderStat = GameStat & {
	skills: number[];
};

// Nhúng luôn patch/server/rankBucket của meta để đứng độc lập được.
// Không kèm các nhóm build: tier list trả hàng trăm dòng nên chúng sẽ chiếm ~85% payload.
export type ChampionStatSummary = {
	patch: string;
	server: MetaServer;
	rankBucket: RankBucket;

	position: Position;
	championId: number;
	championSlug: string;

	games: number;
	wins: number;
	winRate: number;
	pickRate: number;
	bans: number;
	banRate: number;
	tierScore: number;
	tier: ChampionTier;

	avgKills: number;
	avgDeaths: number;
	avgAssists: number;
	avgKda: number;
	avgKp: number;
	avgCsPerMin: number;
	avgGoldPerMin: number;
	avgDmgPerMin: number;
	avgDmgTakenPerMin: number;
	avgCcPerMin: number;
	avgPhysicalDmg: number;
	avgMagicDmg: number;
	avgTrueDmg: number;
	avgPenta: number;
	avgSoloKills: number;
	avgPerfScore: number;

	bestMatchUps: MatchupStat[];
};

// summary + build, cho endpoint đọc 1 tướng.
export type ChampionStat = ChampionStatSummary & {
	bestSpellCombos: SpellComboStat[];
	bestRunes: RuneStat[];
	bestLegendaryItems: ItemStat[];
	bestBootItems: ItemStat[];
	bestStarterSets: ItemSetStat[];
	bestSkillsLeveled: SkillOrderStat[];
	bestFirstLegendItems: ItemStat[];
	bestFirstThreeItems: ItemSetStat[];
};

export type Meta = {
	id: string;
	patch: string;
	server: MetaServer;
	rankBucket: RankBucket;
	totalMatches: number;
	updatedAt: string;
	championStats: ChampionStatSummary[];
};
