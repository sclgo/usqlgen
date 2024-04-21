package lang

import "github.com/samber/lo"

func IfEmpty[T comparable](value, alt T) T {
	result, _ := lo.Coalesce(value, alt)
	return result
}
