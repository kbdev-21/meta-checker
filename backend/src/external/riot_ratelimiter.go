package external

import (
	"context"
	"sync"
	"time"
)

// Riot đếm rate limit theo HOST, mỗi host một budget riêng:
//   - platform (vn2, kr, euw1...) cho league-v4 / summoner-v4  => 15 budget
//   - region (americas, asia, europe, sea) cho account-v1 / match-v5 => 4 budget dùng chung
//
// Limiter giữ nhịp để luôn chạy sát trần mà không vượt. Nó KHÔNG phủ method rate limit
// (limit riêng của từng endpoint); với key 20 req/s + 100 req/2min thì app limit luôn
// chạm trước nên tạm đủ.

// Chạy dưới trần thật 10%: bộ đếm tự giữ không bao giờ khớp tuyệt đối với bộ đếm của Riot
// (lệch đồng hồ, request đang bay), chừa đệm để không ăn 429 vì sai số.
const riotRateSafetyPercent = 90

// Mức ưu tiên khi xếp hàng. Request của user chen trước crawler, nếu không thì crawler
// chạy càng khoẻ, web càng chậm.
type Priority int

const (
	PriorityHigh Priority = iota // request do user kích hoạt
	PriorityLow                  // crawler chạy nền
)

// Nhịp kiểm tra lại của PriorityLow khi đang nhường chỗ cho PriorityHigh.
const lowPriorityRetryInterval = 20 * time.Millisecond

// Số lần thử lại tối đa khi Riot trả 429.
const riotMaxRetries = 3

type ctxPriorityKey struct{}

// Gắn mức ưu tiên vào ctx. RiotClient đọc ra lúc xếp hàng.
func WithPriority(ctx context.Context, p Priority) context.Context {
	return context.WithValue(ctx, ctxPriorityKey{}, p)
}

// Mặc định PriorityHigh: quên gắn thì bị đối xử như request của user, chậm crawler
// chứ không chậm web.
func priorityOf(ctx context.Context) Priority {
	p, ok := ctx.Value(ctxPriorityKey{}).(Priority)
	if !ok {
		return PriorityHigh
	}
	return p
}

// ---------- limiter ----------

type riotRateLimiter struct {
	mu      sync.Mutex
	hosts   map[string]*hostLimiter
	perSec  int
	per2Min int
}

func newRiotRateLimiter(perSec, per2Min int) *riotRateLimiter {
	return &riotRateLimiter{
		hosts:   map[string]*hostLimiter{},
		perSec:  perSec,
		per2Min: per2Min,
	}
}

// Chặn tới khi host còn slot, hoặc ctx hết hạn.
func (r *riotRateLimiter) acquire(ctx context.Context, host string, p Priority) error {
	return r.hostLimiter(host).acquire(ctx, p)
}

// ---------- private ----------

func (r *riotRateLimiter) hostLimiter(host string) *hostLimiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	l, ok := r.hosts[host]
	if !ok {
		l = &hostLimiter{
			second:     newRateWindow(r.perSec, time.Second),
			twoMinutes: newRateWindow(r.per2Min, 2*time.Minute),
		}
		r.hosts[host] = l
	}
	return l
}

// Budget của 1 host: phải qua cả hai cửa sổ mới được gửi.
type hostLimiter struct {
	mu          sync.Mutex
	second      rateWindow
	twoMinutes  rateWindow
	highWaiting int // số PriorityHigh đang xếp hàng
}

func (l *hostLimiter) acquire(ctx context.Context, p Priority) error {
	if p == PriorityHigh {
		l.mu.Lock()
		l.highWaiting++
		l.mu.Unlock()
		defer func() {
			l.mu.Lock()
			l.highWaiting--
			l.mu.Unlock()
		}()
	}

	for {
		l.mu.Lock()
		now := time.Now()
		wait := max(l.second.wait(now), l.twoMinutes.wait(now))
		// Còn slot nhưng đang có request của user chờ thì nhường.
		yielding := p == PriorityLow && l.highWaiting > 0
		if wait == 0 && !yielding {
			l.second.record(now)
			l.twoMinutes.record(now)
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()

		if yielding {
			wait = lowPriorityRetryInterval
		}
		err := sleepCtx(ctx, wait)
		if err != nil {
			return err
		}
	}
}

// Cửa sổ trượt: giữ mốc thời gian của tối đa limit request gần nhất.
type rateWindow struct {
	limit  int
	period time.Duration
	hits   []time.Time
}

func newRateWindow(limit int, period time.Duration) rateWindow {
	return rateWindow{limit: max(1, limit*riotRateSafetyPercent/100), period: period}
}

// 0 = còn slot. Khác 0 = phải đợi bấy lâu nữa mới có slot.
func (w *rateWindow) wait(now time.Time) time.Duration {
	cutoff := now.Add(-w.period)
	i := 0
	for i < len(w.hits) && !w.hits[i].After(cutoff) {
		i++
	}
	w.hits = w.hits[i:]

	if len(w.hits) < w.limit {
		return 0
	}
	return w.hits[0].Add(w.period).Sub(now)
}

func (w *rateWindow) record(now time.Time) {
	w.hits = append(w.hits, now)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
