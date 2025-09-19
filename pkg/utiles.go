package pkg

import "strings"

func TextOnly(s string) string {
	// strip emoji variation selectors so we render text glyphs
	return strings.Map(func(r rune) rune {
		if r == 0xFE0F || r == 0xFE0E {
			return -1
		}
		return r
	}, s)
}

func EllipsizeMiddle(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	keep := max - 1
	left := keep / 2
	right := keep - left
	return string(r[:left]) + "…" + string(r[len(r)-right:])
}
