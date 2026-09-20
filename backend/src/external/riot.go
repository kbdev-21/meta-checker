package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Số match gọi song song mỗi lượt trong GetMatchesByIds, tránh bị rate limit.
const matchFetchBatchSize = 4

// region: americas | asia | europe | sea (account-v1, match-v5)
// platform: vn2 | kr | na1 | euw1 | ... (league-v4, summoner-v4)
type RiotClient struct {
	apiKey string
	http   *http.Client
}

func NewRiotClient(apiKey string) *RiotClient {
	return &RiotClient{apiKey: apiKey, http: &http.Client{Timeout: 10 * time.Second}}
}

// Riot trả status khác 200.
type RiotError struct {
	Path       string
	StatusCode int
	Body       string
}

func (e *RiotError) Error() string {
	return fmt.Sprintf("riot api %s: %d %s", e.Path, e.StatusCode, e.Body)
}

// Riot trả 404: resource (account, summoner...) không tồn tại.
func IsRiotNotFound(err error) bool {
	var re *RiotError
	return errors.As(err, &re) && re.StatusCode == http.StatusNotFound
}

func (c *RiotClient) get(ctx context.Context, host, path string, query url.Values, out any) error {
	u := "https://" + host + ".api.riotgames.com" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Riot-Token", c.apiKey)

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return &RiotError{Path: path, StatusCode: res.StatusCode, Body: string(body)}
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// ---------- account-v1 ----------

func (c *RiotClient) GetAccountByRiotId(ctx context.Context, region, gameName, tagLine string) (*AccountDto, error) {
	var out AccountDto
	path := "/riot/account/v1/accounts/by-riot-id/" + url.PathEscape(gameName) + "/" + url.PathEscape(tagLine)
	return &out, c.get(ctx, region, path, nil, &out)
}

func (c *RiotClient) GetAccountByPuuid(ctx context.Context, region, puuid string) (*AccountDto, error) {
	var out AccountDto
	return &out, c.get(ctx, region, "/riot/account/v1/accounts/by-puuid/"+puuid, nil, &out)
}

// ---------- league-v4 ----------

func (c *RiotClient) GetLeagueEntriesByPuuid(ctx context.Context, platform, puuid string) ([]LeagueEntryDto, error) {
	var out []LeagueEntryDto
	err := c.get(ctx, platform, "/lol/league/v4/entries/by-puuid/"+puuid, nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RiotClient) GetChallengerLeague(ctx context.Context, platform string, queue QueueType) (*LeagueListDto, error) {
	var out LeagueListDto
	return &out, c.get(ctx, platform, "/lol/league/v4/challengerleagues/by-queue/"+string(queue), nil, &out)
}

func (c *RiotClient) GetGrandmasterLeague(ctx context.Context, platform string, queue QueueType) (*LeagueListDto, error) {
	var out LeagueListDto
	return &out, c.get(ctx, platform, "/lol/league/v4/grandmasterleagues/by-queue/"+string(queue), nil, &out)
}

func (c *RiotClient) GetMasterLeague(ctx context.Context, platform string, queue QueueType) (*LeagueListDto, error) {
	var out LeagueListDto
	return &out, c.get(ctx, platform, "/lol/league/v4/masterleagues/by-queue/"+string(queue), nil, &out)
}

// ---------- match-v5 ----------

func (c *RiotClient) GetMatchIdsByPuuid(ctx context.Context, region, puuid string, opts MatchIdsOptions) ([]string, error) {
	q := url.Values{}
	if opts.StartTime > 0 {
		q.Set("startTime", strconv.FormatInt(opts.StartTime, 10))
	}
	if opts.EndTime > 0 {
		q.Set("endTime", strconv.FormatInt(opts.EndTime, 10))
	}
	if opts.Queue != nil {
		q.Set("queue", strconv.Itoa(*opts.Queue))
	}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Start > 0 {
		q.Set("start", strconv.Itoa(opts.Start))
	}
	if opts.Count > 0 {
		q.Set("count", strconv.Itoa(opts.Count))
	}

	var out []string
	err := c.get(ctx, region, "/lol/match/v5/matches/by-puuid/"+puuid+"/ids", q, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *RiotClient) GetMatch(ctx context.Context, region, matchId string) (*MatchDto, error) {
	var out MatchDto
	return &out, c.get(ctx, region, "/lol/match/v5/matches/"+matchId, nil, &out)
}

// Gọi GetMatch cho từng id, mỗi lượt matchFetchBatchSize goroutine song song.
// Kết quả theo thứ tự matchIds. Có match lỗi => trả lỗi, bỏ cả kết quả.
func (c *RiotClient) GetMatchesByIdsInParallel(ctx context.Context, region string, matchIds []string) ([]*MatchDto, error) {
	out := make([]*MatchDto, 0, len(matchIds))
	for start := 0; start < len(matchIds); start += matchFetchBatchSize {
		batch := matchIds[start:min(start+matchFetchBatchSize, len(matchIds))]
		fetched := make([]*MatchDto, len(batch))
		errs := make([]error, len(batch))
		var wg sync.WaitGroup
		for i, id := range batch {
			wg.Go(func() {
				fetched[i], errs[i] = c.GetMatch(ctx, region, id)
			})
		}
		wg.Wait()
		err := errors.Join(errs...)
		if err != nil {
			return nil, err
		}
		out = append(out, fetched...)
	}
	return out, nil
}

// ---------- summoner-v4 ----------

func (c *RiotClient) GetSummonerByPuuid(ctx context.Context, platform, puuid string) (*SummonerDto, error) {
	var out SummonerDto
	return &out, c.get(ctx, platform, "/lol/summoner/v4/summoners/by-puuid/"+puuid, nil, &out)
}
