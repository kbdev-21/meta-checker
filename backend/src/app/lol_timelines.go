package app

import (
	"context"
	"slices"

	"backend/src/external"
)

// Item mua trước mốc này (ms từ đầu trận) tính là starter set.
const starterSetWindowMs = 60_000

// ---------- private ----------

// Build của 1 participant tách từ timeline. Slice luôn khác nil: cột NOT NULL, slice nil pgx ghi thành NULL.
type timelineBuild struct {
	StarterSets          []int32
	SkillsLeveled        []int32
	LegendItemsPurchased []int32
}

func emptyTimelineBuild() timelineBuild {
	return timelineBuild{StarterSets: []int32{}, SkillsLeveled: []int32{}, LegendItemsPurchased: []int32{}}
}

// Fetch timeline (bằng riotTimeline, key thứ 2) cho mọi id, không cần chờ MatchDto để biết mode.
// Kết quả: match id => timeline; Riot 404 => không có key.
func (a *Application) fetchTimelinesOf(ctx context.Context, region string, ids []string) (map[string]*external.MatchTimelineDto, error) {
	timelines, err := a.riotTimeline.FetchMatchTimelinesByIdsInParallel(ctx, region, ids)
	if err != nil {
		return nil, err
	}
	out := map[string]*external.MatchTimelineDto{}
	for i, id := range ids {
		if timelines[i] != nil {
			out[id] = timelines[i]
		}
	}
	return out, nil
}

// item id => ItemType, từ lol_items (sync ddragon). Item không có trong bảng => không có key.
func (a *Application) itemTypesById(ctx context.Context) (map[int32]ItemType, error) {
	items, err := a.q.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[int32]ItemType, len(items))
	for _, it := range items {
		out[it.ID] = ItemType(it.Type)
	}
	return out, nil
}

type timelinePurchase struct {
	itemId    int32
	timestamp int64
}

// Tách timeline thành build của từng participant, key = participantId (1..10).
// Puuid trong timeline mã hóa theo key khác nên luôn map qua participantId.
// timeline nil (Riot 404) => build rỗng cho cả 10 người.
func timelineBuildsOf(timeline *external.MatchTimelineDto, itemTypes map[int32]ItemType) map[int]timelineBuild {
	purchases := map[int][]timelinePurchase{}
	skills := map[int][]int32{}
	frames := []external.TimelineFrameDto{}
	if timeline != nil {
		frames = timeline.Info.Frames
	}
	for _, f := range frames {
		for _, e := range f.Events {
			switch e.Type {
			case "ITEM_PURCHASED":
				purchases[e.ParticipantId] = append(purchases[e.ParticipantId], timelinePurchase{itemId: int32(e.ItemId), timestamp: e.Timestamp})
			case "ITEM_UNDO":
				// beforeId = item bị hoàn tác; 0 = hoàn tác lượt bán, không đụng tới lượt mua.
				if e.BeforeId != 0 {
					purchases[e.ParticipantId] = removeLastPurchase(purchases[e.ParticipantId], int32(e.BeforeId))
				}
			case "SKILL_LEVEL_UP":
				// EVOLVE (Kha'Zix, Kai'Sa...) là tiến hóa skill, không phải điểm skill.
				if e.LevelUpType == "NORMAL" {
					skills[e.ParticipantId] = append(skills[e.ParticipantId], int32(e.SkillSlot))
				}
			}
		}
	}

	// Riot có bắn vài event participantId 0 ở giây 0, không thuộc người chơi nào: chỉ lấy 1..10.
	out := map[int]timelineBuild{}
	for participantId := 1; participantId <= 10; participantId++ {
		build := emptyTimelineBuild()
		build.SkillsLeveled = append(build.SkillsLeveled, skills[participantId]...)
		for _, p := range purchases[participantId] {
			itemType := itemTypes[p.itemId]
			if p.timestamp < starterSetWindowMs && itemType != ItemTypeTrinket {
				build.StarterSets = append(build.StarterSets, p.itemId)
			}
			// Bán rồi mua lại cùng món thì chỉ giữ lần đầu.
			if itemType == ItemTypeLegendary && !slices.Contains(build.LegendItemsPurchased, p.itemId) {
				build.LegendItemsPurchased = append(build.LegendItemsPurchased, p.itemId)
			}
		}
		// sort để cùng bộ đồ khởi đầu thì GROUP BY ra cùng một mảng.
		slices.Sort(build.StarterSets)
		out[participantId] = build
	}
	return out
}

// Xóa lượt mua gần nhất của itemId (ITEM_UNDO hoàn tác lượt mua mới nhất). Không có => giữ nguyên.
func removeLastPurchase(purchases []timelinePurchase, itemId int32) []timelinePurchase {
	for i := len(purchases) - 1; i >= 0; i-- {
		if purchases[i].itemId == itemId {
			return slices.Delete(purchases, i, i+1)
		}
	}
	return purchases
}
