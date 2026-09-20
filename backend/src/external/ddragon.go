package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const ddragonBase = "https://ddragon.leagueoflegends.com"

// lang: en_US | vi_VN | ko_KR | ...
type DDragonClient struct {
	http *http.Client
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

func (c *DDragonClient) GetCurrentVersion(ctx context.Context) (string, error) {
	var versions []string
	err := c.get(ctx, "/api/versions.json", &versions)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", errors.New("ddragon: empty versions")
	}
	return versions[0], nil
}

// key = champion id dạng tên (vd "Aatrox"); DDChampion.Key = championId số trong match.
func (c *DDragonClient) GetChampions(ctx context.Context, version, lang string) (map[string]DDChampion, error) {
	var out ddResponse[DDChampion]
	err := c.get(ctx, dataPath(version, lang, "champion.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

// key = item id (vd "3031").
func (c *DDragonClient) GetItems(ctx context.Context, version, lang string) (map[string]DDItem, error) {
	var out ddResponse[DDItem]
	err := c.get(ctx, dataPath(version, lang, "item.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

// key = spell id dạng tên (vd "SummonerFlash"); DDSummonerSpell.Key = summoner1Id/2Id trong match.
func (c *DDragonClient) GetSummonerSpells(ctx context.Context, version, lang string) (map[string]DDSummonerSpell, error) {
	var out ddResponse[DDSummonerSpell]
	err := c.get(ctx, dataPath(version, lang, "summoner.json"), &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *DDragonClient) GetRunes(ctx context.Context, version, lang string) ([]DDRuneTree, error) {
	var out []DDRuneTree
	err := c.get(ctx, dataPath(version, lang, "runesReforged.json"), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
