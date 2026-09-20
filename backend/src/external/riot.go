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

// Số match gọi song song mỗi lượt trong GetMatchesByIdsInParallel. Việc giữ nhịp rate limit do
// riotRateLimiter lo, hằng này chỉ còn giới hạn số goroutine cùng chờ.
const matchFetchBatchSize = 10

// region: americas | asia | europe | sea (account-v1, match-v5)
// platform: vn2 | kr | na1 | euw1 | ... (league-v4, summoner-v4)
type RiotClient struct {
	apiKey  string
	http    *http.Client
	limiter *riotRateLimiter
}

// perSec / per2Min là trần THẬT của key; limiter tự chừa đệm an toàn bên dưới.
func NewRiotClient(apiKey string, perSec, per2Min int) *RiotClient {
	return &RiotClient{
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
		limiter: newRiotRateLimiter(perSec, per2Min),
	}
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

// Xếp hàng rate limit theo host trước khi gửi. Riot vẫn trả 429 được (lệch bộ đếm, hoặc
// 429 ở tầng service của Riot) nên tôn trọng Retry-After và thử lại tối đa riotMaxRetries lần.
func (c *RiotClient) get(ctx context.Context, host, path string, query url.Values, out any) error {
	priority := priorityOf(ctx)
	for attempt := 0; ; attempt++ {
		err := c.limiter.acquire(ctx, host, priority)
		if err != nil {
			return err
		}

		res, err := c.send(ctx, host, path, query)
		if err != nil {
			return err
		}

		if res.StatusCode == http.StatusTooManyRequests && attempt < riotMaxRetries {
			retryAfter := retryAfterOf(res)
			res.Body.Close()
			err = sleepCtx(ctx, retryAfter)
			if err != nil {
				return err
			}
			continue
		}

		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			return &RiotError{Path: path, StatusCode: res.StatusCode, Body: string(body)}
		}
		err = json.NewDecoder(res.Body).Decode(out)
		res.Body.Close()
		return err
	}
}

func (c *RiotClient) send(ctx context.Context, host, path string, query url.Values) (*http.Response, error) {
	u := "https://" + host + ".api.riotgames.com" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Riot-Token", c.apiKey)
	return c.http.Do(req)
}

// Riot trả Retry-After theo giây. Thiếu / không parse được thì đợi 1 giây.
func retryAfterOf(res *http.Response) time.Duration {
	sec, err := strconv.Atoi(res.Header.Get("Retry-After"))
	if err != nil || sec <= 0 {
		return time.Second
	}
	return time.Duration(sec) * time.Second
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
