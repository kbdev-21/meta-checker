package shared

import (
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func ToInt32s(ss []string) ([]int32, error) {
	out := make([]int32, 0, len(ss))
	for _, s := range ss {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		out = append(out, int32(n))
	}
	return out, nil
}

// Chữ thường + bỏ dấu, vd "Nguyễn Văn A" => "nguyen van a".
// Tách ký tự thành chữ gốc + dấu (NFD), bỏ các dấu (Mn), rồi ghép lại (NFC).
// "đ" không phải chữ gốc + dấu nên phải thay tay.
func NormalizeString(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "đ", "d")

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(out)
}
