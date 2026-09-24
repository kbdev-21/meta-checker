package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const ddragonBase = "https://ddragon.leagueoflegends.com"

// Patch cỡ 2 tuần mới đổi nên TTL 1h là thừa tươi; trùng nhịp sched chạy mỗi giờ.
const ddragonVersionTTL = time.Hour

// lang: en_US | vi_VN | ko_KR | ...
type DDragonClient struct {
	http *http.Client

	// Cache version: mọi request đọc analytics đều cần, không lẽ gọi ddragon mỗi lần.
	versionMu      sync.Mutex
	version        string
	versionExpires time.Time
}

func NewDDragonClient() *DDragonClient {
	return &DDragonClient{http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *DDragonClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ddragonBase+path, nil)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("ddragon %s: %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func dataPath(version, lang, file string) string {
	return "/cdn/" + version + "/data/" + lang + "/" + file
}

// vd: https://ddragon.leagueoflegends.com/cdn/16.18.1/img/champion/Aatrox.png
func DDImgUrl(version string, img DDImage) string {
	return ddragonBase + "/cdn/" + version + "/img/" + img.Group + "/" + img.Full
}

// Icon của rune và cây rune nằm ở path riêng, KHÔNG kèm version như DDImgUrl.
// vd: https://ddragon.leagueoflegends.com/cdn/img/perk-images/Styles/Domination/Electrocute/Electrocute.png
func DDRuneIconUrl(icon string) string {
	return ddragonBase + "/cdn/img/" + icon
}

// Cache 1h. Giữ lock qua cả cú fetch: nhiều request cùng hết hạn một lúc thì chỉ 1 cái
// đi gọi ddragon, số còn lại chờ rồi ăn cache mới.
func (c *DDragonClient) FetchCurrentVersion(ctx context.Context) (string, error) {
	c.versionMu.Lock()
	defer c.versionMu.Unlock()

	if c.version != "" && time.Now().Before(c.versionExpires) {
		return c.version, nil
	}

	var versions []string
	err := c.get(ctx, "/api/versions.json", &versions)
	if err == nil && len(versions) == 0 {
		err = errors.New("ddragon: empty versions")
	}
	if err != nil {
		// Còn bản cũ thì xài tạm: version cũ vẫn dùng được, hỏng cả request thì không.
		if c.version != "" {
			return c.version, nil
		}
		return "", err
	}

	c.version = versions[0]
	c.versionExpires = time.Now().Add(ddragonVersionTTL)
	return c.version, nil
}

// key = champion id dạng tên (vd "Aatrox"); DDChampion.Key = championId số trong match.
func (c *DDragonClient) FetchChampions(ctx context.Context, version, lang string) (map[string]DDChampion, error) {
	var out ddResponse[DDChampion]
	err := c.get(ctx, dataPath(version, lang, "champion.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

// key = item id (vd "3031").
func (c *DDragonClient) FetchItems(ctx context.Context, version, lang string) (map[string]DDItem, error) {
	var out ddResponse[DDItem]
	err := c.get(ctx, dataPath(version, lang, "item.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

// key = spell id dạng tên (vd "SummonerFlash"); DDSummonerSpell.Key = summoner1Id/2Id trong match.
func (c *DDragonClient) FetchSummonerSpells(ctx context.Context, version, lang string) (map[string]DDSummonerSpell, error) {
	var out ddResponse[DDSummonerSpell]
	err := c.get(ctx, dataPath(version, lang, "summoner.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *DDragonClient) FetchRunes(ctx context.Context, version, lang string) ([]DDRuneTree, error) {
	var out []DDRuneTree
	err := c.get(ctx, dataPath(version, lang, "runesReforged.json"), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
