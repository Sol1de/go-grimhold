package cli

import (
	"fmt"
	"strings"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

func pick(highIsGood bool, ratio float64) string {
	if !highIsGood {
		ratio = 1 - ratio
	}
	switch {
	case ratio > 0.5:
		return green
	case ratio > 0.25:
		return yellow
	}
	return red
}

func hpBar(hp, maxHP int) string {
	const width = 10
	hp = max(hp, 0)
	filled := min(hp*width/max(maxHP, 1), width)
	return fmt.Sprintf("%s[%s%s] %d/%d%s", pick(true, float64(hp)/float64(max(maxHP, 1))),
		strings.Repeat("#", filled), strings.Repeat("-", width-filled), hp, maxHP, reset)
}

func danger(d int) string {
	return fmt.Sprintf("%sdanger %d%s", pick(false, float64(d)/4), d, reset)
}
