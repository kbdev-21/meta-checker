package shared

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Nullable[T any] struct {
	Value  T
	IsNull bool
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.IsNull = true
		var v T
		n.Value = v
		return nil
	}

	err := json.Unmarshal(data, &n.Value)
	if err != nil {
		return err
	}
	n.IsNull = false
	return nil
}

// pgtype <-> Nullable

func NullableText[T ~string](t pgtype.Text) Nullable[T] {
	return Nullable[T]{Value: T(t.String), IsNull: !t.Valid}
}

func PgText[T ~string](n Nullable[T]) pgtype.Text {
	return pgtype.Text{String: string(n.Value), Valid: !n.IsNull}
}

func NullableInt4(i pgtype.Int4) Nullable[int32] {
	return Nullable[int32]{Value: i.Int32, IsNull: !i.Valid}
}

func PgInt4(n Nullable[int32]) pgtype.Int4 {
	return pgtype.Int4{Int32: n.Value, Valid: !n.IsNull}
}

func NullableTime(t pgtype.Timestamptz) Nullable[time.Time] {
	return Nullable[time.Time]{Value: t.Time, IsNull: !t.Valid}
}
