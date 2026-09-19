package shared

import "strconv"

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
