package app

import (
	"context"
	"slices"
	"strconv"
	"time"

	"backend/src/db"
	"backend/src/external"
	"backend/src/shared"
)

const ddragonLang = "en_US"

// ---------- enum ----------

type ItemType string

const (
	ItemTypeConsumable ItemType = "CONSUMABLE" // tags có Consumable
	ItemTypeTrinket    ItemType = "TRINKET"    // tags có Trinket
	ItemTypeBoots      ItemType = "BOOTS"      // tags có Boots
	ItemTypeStarter    ItemType = "STARTER"    // không from, không into, còn mua được
	ItemTypeBasic      ItemType = "BASIC"      // không from, có into
	ItemTypeEpic       ItemType = "EPIC"       // có from, có into
	ItemTypeLegendary  ItemType = "LEGENDARY"  // có from, không into
)

// ---------- entity ----------

type LolChampion struct {
	Id        int32     `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Title     string    `json:"title"`
	ImgUrl    string    `json:"imgUrl"`
	Patch     string    `json:"patch"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToLolChampion(c db.LolChampion) LolChampion {
	return LolChampion{
		Id:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		Title:     c.Title,
		ImgUrl:    c.ImgUrl,
		Patch:     c.Patch,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

type LolItem struct {
	Id        int32     `json:"id"`
	Name      string    `json:"name"`
	Plaintext string    `json:"plaintext"`
	Type      ItemType  `json:"type"`
	GoldTotal int32     `json:"goldTotal"`
	FromItems []int32   `json:"fromItems"`
	IntoItems []int32   `json:"intoItems"`
	IsSr      bool      `json:"isSr"`
	ImgUrl    string    `json:"imgUrl"`
	Patch     string    `json:"patch"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToLolItem(i db.LolItem) LolItem {
	return LolItem{
		Id:        i.ID,
		Name:      i.Name,
		Plaintext: i.Plaintext,
		Type:      ItemType(i.Type),
		GoldTotal: i.GoldTotal,
		FromItems: i.FromItems,
		IntoItems: i.IntoItems,
		IsSr:      i.IsSr,
		ImgUrl:    i.ImgUrl,
		Patch:     i.Patch,
		UpdatedAt: i.UpdatedAt.Time,
	}
}

// ---------- logic ----------

func (a *Application) GetLolChampions(ctx context.Context) ([]LolChampion, error) {
	rows, err := a.q.ListChampions(ctx)
	if err != nil {
		return nil, err
	}
	out := []LolChampion{}
	for _, r := range rows {
		out = append(out, ToLolChampion(r))
	}
	return out, nil
}

func (a *Application) GetLolItems(ctx context.Context) ([]LolItem, error) {
	rows, err := a.q.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	out := []LolItem{}
	for _, r := range rows {
		out = append(out, ToLolItem(r))
	}
	return out, nil
}

func (a *Application) UpdateLolChampions(ctx context.Context) error {
	version, err := a.ddragon.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}
	champs, err := a.ddragon.GetChampions(ctx, version, ddragonLang)
	if err != nil {
		return err
	}

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)

	for _, c := range champs {
		id, err := strconv.Atoi(c.Key)
		if err != nil {
			return err
		}
		err = q.UpsertChampion(ctx, db.UpsertChampionParams{
			ID:     int32(id),
			Slug:   c.Id,
			Name:   c.Name,
			Title:  c.Title,
			ImgUrl: external.DDImgUrl(version, c.Image),
			Patch:  shared.PatchOf(version),
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (a *Application) UpdateLolItems(ctx context.Context) error {
	version, err := a.ddragon.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}
	items, err := a.ddragon.GetItems(ctx, version, ddragonLang)
	if err != nil {
		return err
	}

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)

	for key, it := range items {
		t, ok := itemTypeOf(it)
		if !ok {
			continue
		}
		id, err := strconv.Atoi(key)
		if err != nil {
			return err
		}
		from, err := shared.ToInt32s(it.From)
		if err != nil {
			return err
		}
		into, err := shared.ToInt32s(it.Into)
		if err != nil {
			return err
		}
		err = q.UpsertItem(ctx, db.UpsertItemParams{
			ID:        int32(id),
			Name:      it.Name,
			Plaintext: it.Plaintext,
			Type:      string(t),
			GoldTotal: int32(it.Gold.Total),
			FromItems: from,
			IntoItems: into,
			IsSr:      it.Maps["11"],
			ImgUrl:    external.DDImgUrl(version, it.Image),
			Patch:     shared.PatchOf(version),
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ---------- private ----------

// Xét theo thứ tự các ItemType ở trên. ok = false nếu không thuộc loại nào (item không mua được, item riêng của tướng...).
func itemTypeOf(it external.DDItem) (t ItemType, ok bool) {
	hasFrom, hasInto := len(it.From) > 0, len(it.Into) > 0
	switch {
	case slices.Contains(it.Tags, "Consumable"):
		return ItemTypeConsumable, true
	case slices.Contains(it.Tags, "Trinket"):
		return ItemTypeTrinket, true
	case slices.Contains(it.Tags, "Boots"):
		return ItemTypeBoots, true
	case !hasFrom && !hasInto && it.Gold.Purchasable:
		return ItemTypeStarter, true
	case !hasFrom && hasInto:
		return ItemTypeBasic, true
	case hasFrom && hasInto:
		return ItemTypeEpic, true
	case hasFrom && !hasInto:
		return ItemTypeLegendary, true
	}
	return "", false
}
