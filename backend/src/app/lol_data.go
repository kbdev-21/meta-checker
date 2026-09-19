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

type Champion struct {
	Id        int32     `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Title     string    `json:"title"`
	ImgUrl    string    `json:"imgUrl"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToChampion(c db.Champion) Champion {
	return Champion{
		Id:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		Title:     c.Title,
		ImgUrl:    c.ImgUrl,
		Version:   c.Version,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

type Item struct {
	Id        int32     `json:"id"`
	Name      string    `json:"name"`
	Plaintext string    `json:"plaintext"`
	Type      ItemType  `json:"type"`
	GoldTotal int32     `json:"goldTotal"`
	FromItems []int32   `json:"fromItems"`
	IntoItems []int32   `json:"intoItems"`
	IsSr      bool      `json:"isSr"`
	ImgUrl    string    `json:"imgUrl"`
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToItem(i db.Item) Item {
	return Item{
		Id:        i.ID,
		Name:      i.Name,
		Plaintext: i.Plaintext,
		Type:      ItemType(i.Type),
		GoldTotal: i.GoldTotal,
		FromItems: i.FromItems,
		IntoItems: i.IntoItems,
		IsSr:      i.IsSr,
		ImgUrl:    i.ImgUrl,
		Version:   i.Version,
		UpdatedAt: i.UpdatedAt.Time,
	}
}

// ---------- logic ----------

func (a *Application) GetChampions(ctx context.Context) ([]Champion, error) {
	rows, err := a.q.ListChampions(ctx)
	if err != nil {
		return nil, err
	}
	out := []Champion{}
	for _, r := range rows {
		out = append(out, ToChampion(r))
	}
	return out, nil
}

func (a *Application) GetItems(ctx context.Context) ([]Item, error) {
	rows, err := a.q.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	out := []Item{}
	for _, r := range rows {
		out = append(out, ToItem(r))
	}
	return out, nil
}

func (a *Application) UpsertChampions(ctx context.Context) error {
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
			ID:      int32(id),
			Slug:    c.Id,
			Name:    c.Name,
			Title:   c.Title,
			ImgUrl:  external.DDImgUrl(version, c.Image),
			Version: version,
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (a *Application) UpsertItems(ctx context.Context) error {
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
			Version:   version,
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
