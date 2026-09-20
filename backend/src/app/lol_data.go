package app

import (
	"context"
	"strconv"
	"time"

	"backend/src/db"
	"backend/src/external"
	"backend/src/shared"

	"github.com/jackc/pgx/v5/pgtype"
)

const ddragonLang = "en_US"

// ---------- entity ----------

type Champion struct {
	Id        int32     `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Title     string    `json:"title"`
	ImgUrl    string    `json:"imgUrl"`
	Patch     string    `json:"patch"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToChampion(c db.LolChampion) Champion {
	return Champion{
		Id:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		Title:     c.Title,
		ImgUrl:    c.ImgUrl,
		Patch:     c.Patch,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

type Item struct {
	Id              int32     `json:"id"`
	Name            string    `json:"name"`
	Plaintext       string    `json:"plaintext"`
	Type            ItemType  `json:"type"`
	GoldTotal       int32     `json:"goldTotal"`
	FromItems       []int32   `json:"fromItems"`
	IntoItems       []int32   `json:"intoItems"`
	IsSummonersRift bool      `json:"isSummonersRift"`
	ImgUrl          string    `json:"imgUrl"`
	Patch           string    `json:"patch"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func ToItem(i db.LolItem) Item {
	return Item{
		Id:              i.ID,
		Name:            i.Name,
		Plaintext:       i.Plaintext,
		Type:            ItemType(i.Type),
		GoldTotal:       i.GoldTotal,
		FromItems:       i.FromItems,
		IntoItems:       i.IntoItems,
		IsSummonersRift: i.IsSummonersRift,
		ImgUrl:          i.ImgUrl,
		Patch:           i.Patch,
		UpdatedAt:       i.UpdatedAt.Time,
	}
}

type Spell struct {
	Id          int32     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImgUrl      string    `json:"imgUrl"`
	Patch       string    `json:"patch"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func ToSpell(s db.LolSpell) Spell {
	return Spell{
		Id:          s.ID,
		Slug:        s.Slug,
		Name:        s.Name,
		Description: s.Description,
		ImgUrl:      s.ImgUrl,
		Patch:       s.Patch,
		UpdatedAt:   s.UpdatedAt.Time,
	}
}

// Gộp cả cây rune lẫn rune. StyleId IsNull => dòng này là một cây.
type Rune struct {
	Id        int32                  `json:"id"`
	StyleId   shared.Nullable[int32] `json:"styleId"`
	Slot      shared.Nullable[int32] `json:"slot"`
	Slug      string                 `json:"slug"`
	Name      string                 `json:"name"`
	ShortDesc string                 `json:"shortDesc"`
	ImgUrl    string                 `json:"imgUrl"`
	Patch     string                 `json:"patch"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

func ToRune(r db.LolRune) Rune {
	return Rune{
		Id:        r.ID,
		StyleId:   shared.NullableInt4(r.StyleID),
		Slot:      shared.NullableInt4(r.Slot),
		Slug:      r.Slug,
		Name:      r.Name,
		ShortDesc: r.ShortDesc,
		ImgUrl:    r.ImgUrl,
		Patch:     r.Patch,
		UpdatedAt: r.UpdatedAt.Time,
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

// version đầy đủ của ddragon ("16.18.1") chỉ dùng để dựng URL ảnh; thứ lưu vào DB là patch ("16.18").
func (a *Application) GetSpells(ctx context.Context) ([]Spell, error) {
	rows, err := a.q.ListSpells(ctx)
	if err != nil {
		return nil, err
	}
	out := []Spell{}
	for _, r := range rows {
		out = append(out, ToSpell(r))
	}
	return out, nil
}

// Cây trước, rồi rune của từng cây theo thứ tự hàng.
func (a *Application) GetRunes(ctx context.Context) ([]Rune, error) {
	rows, err := a.q.ListRunes(ctx)
	if err != nil {
		return nil, err
	}
	out := []Rune{}
	for _, r := range rows {
		out = append(out, ToRune(r))
	}
	return out, nil
}

func (a *Application) UpdateChampions(ctx context.Context) error {
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

// Xem UpdateChampions về version vs patch.
func (a *Application) UpdateItems(ctx context.Context) error {
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
			ID:              int32(id),
			Name:            it.Name,
			Plaintext:       it.Plaintext,
			Type:            string(t),
			GoldTotal:       int32(it.Gold.Total),
			FromItems:       from,
			IntoItems:       into,
			IsSummonersRift: it.Maps["11"],
			ImgUrl:          external.DDImgUrl(version, it.Image),
			Patch:           shared.PatchOf(version),
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Xem UpdateChampions về version vs patch.
func (a *Application) UpdateSpells(ctx context.Context) error {
	version, err := a.ddragon.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}
	spells, err := a.ddragon.GetSummonerSpells(ctx, version, ddragonLang)
	if err != nil {
		return err
	}

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)

	for _, sp := range spells {
		id, err := strconv.Atoi(sp.Key)
		if err != nil {
			return err
		}
		err = q.UpsertSpell(ctx, db.UpsertSpellParams{
			ID:          int32(id),
			Slug:        sp.Id,
			Name:        sp.Name,
			Description: sp.Description,
			ImgUrl:      external.DDImgUrl(version, sp.Image),
			Patch:       shared.PatchOf(version),
		})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Lưu cả cây rune lẫn rune vào lol_runes. Cây phải upsert trước rune của nó vì style_id
// tham chiếu ngược về chính bảng này.
// Xem UpdateChampions về version vs patch.
func (a *Application) UpdateRunes(ctx context.Context) error {
	version, err := a.ddragon.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}
	trees, err := a.ddragon.GetRunes(ctx, version, ddragonLang)
	if err != nil {
		return err
	}

	tx, err := a.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := a.q.WithTx(tx)
	patch := shared.PatchOf(version)

	for _, t := range trees {
		err = q.UpsertRune(ctx, db.UpsertRuneParams{
			ID:     int32(t.Id),
			Slug:   t.Key,
			Name:   t.Name,
			ImgUrl: external.DDRuneIconUrl(t.Icon),
			Patch:  patch,
			// StyleID / Slot bỏ trống: đây là cây, không thuộc cây nào.
		})
		if err != nil {
			return err
		}
		for slot, s := range t.Slots {
			for _, r := range s.Runes {
				err = q.UpsertRune(ctx, db.UpsertRuneParams{
					ID:        int32(r.Id),
					StyleID:   pgtype.Int4{Int32: int32(t.Id), Valid: true},
					Slot:      pgtype.Int4{Int32: int32(slot), Valid: true},
					Slug:      r.Key,
					Name:      r.Name,
					ShortDesc: r.ShortDesc,
					ImgUrl:    external.DDRuneIconUrl(r.Icon),
					Patch:     patch,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}
