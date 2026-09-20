package app

import (
	"backend/src/external"
	"backend/src/shared"
)

// ---------- const ----------

// Trọng số của từng thành phần. Tổng trọng số = 100 nhưng điểm cuối không clamp:
// death nhiều kéo điểm xuống âm (sàn -5).
const (
	perfDefaultWeight = 5.0
	perfKpWeight      = 50.0
	perfMaxKp         = 0.75 // kill participation đạt mức này là chạm trần điểm KP
	perfWinWeight     = 15.0
	perfDp10mWeight   = 20.0
	perfGpmDiffWeight = 10.0
	perfMaxGpmDiff    = 135.0 // chênh lệch gold/phút để chạm trần điểm gold
)

// Điểm death = (perfDp10mBase - dp10m * perfDp10mSlope) / 10 * perfDp10mWeight.
// Hai hằng số lấy nguyên từ công thức gốc: 0 death chạm trần, 3.5 death/10 phút ra 0 điểm,
// 5 death/10 phút chạm sàn.
const (
	perfDp10mBase  = 11.665
	perfDp10mSlope = 3.33
)

// Bù kill participation cho vị trí ít tham chiến. Vị trí không có trong map thì không bù.
var perfKpBonuses = map[Position]float64{
	PositionTop: 0.125,
	PositionMid: 0.075,
	PositionAdc: 0.075,
}

// ---------- input ----------

// Toàn bộ dữ liệu cần để tính perf score.
type perfScoreInput struct {
	Position          Position
	KillParticipation float64
	Deaths            int
	IsWin             bool
	GoldEarned        int
	LaneOpponentGold  shared.Nullable[int] // IsNull = không có đối thủ cùng vị trí
	DurationSec       int32
}

// ---------- logic ----------

// Pure function: chỉ phụ thuộc perfScoreInput.
func perfScoreOf(in perfScoreInput) int32 {
	// Trận 0 giây (game hỏng) không tính được gold/phút lẫn death/10 phút.
	if in.DurationSec <= 0 {
		return 0
	}
	durationMin := float64(in.DurationSec) / 60

	kpScore := min((in.KillParticipation+perfKpBonuses[in.Position])/perfMaxKp*perfKpWeight, perfKpWeight)

	winScore := 0.0
	if in.IsWin {
		winScore = perfWinWeight
	}

	dp10m := float64(in.Deaths) / durationMin * 10
	dp10mScore := (perfDp10mBase - dp10m*perfDp10mSlope) / 10 * perfDp10mWeight
	dp10mScore = min(max(dp10mScore, -perfDp10mWeight/2), perfDp10mWeight)

	score := perfDefaultWeight + kpScore + winScore + dp10mScore + perfGpmDiffScoreOf(in, durationMin)
	return int32(score)
}

// ---------- private ----------

// Điểm chênh lệch gold/phút với đối thủ cùng vị trí, mốc trung tính là nửa trọng số.
// SPT không đua gold nên luôn trung tính; không có đối thủ cùng vị trí thì cũng trung tính
// thay vì đoán gold của đối thủ.
func perfGpmDiffScoreOf(in perfScoreInput, durationMin float64) float64 {
	neutral := perfGpmDiffWeight / 2.0
	if in.Position == PositionSpt || in.LaneOpponentGold.IsNull {
		return neutral
	}
	gpmDiff := float64(in.GoldEarned-in.LaneOpponentGold.Value) / durationMin
	return min(max(neutral+gpmDiff/perfMaxGpmDiff*neutral, 0), perfGpmDiffWeight)
}

// Gold của đối thủ khác team, cùng vị trí.
// Vị trí UNK (ARAM, Arena...) không có khái niệm đối lane => IsNull.
func laneOpponentGoldOf(p external.ParticipantDto, participants []external.ParticipantDto) shared.Nullable[int] {
	position := positionOf(p.TeamPosition)
	if position == PositionUnk {
		return shared.Nullable[int]{IsNull: true}
	}
	for _, o := range participants {
		if o.TeamId != p.TeamId && positionOf(o.TeamPosition) == position {
			return shared.Nullable[int]{Value: o.GoldEarned}
		}
	}
	return shared.Nullable[int]{IsNull: true}
}
