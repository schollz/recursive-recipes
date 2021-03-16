package recipe

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func FormatCookingRational(num float64) (s string) {
	// round to nearest eight
	wholeNum := math.Floor(num)
	// log.Debug((num - wholeNum) / 8)
	fractionNum := (math.Round((num-wholeNum)*8) / 8) / .125
	// log.Debug(wholeNum, fractionNum)
	if wholeNum > 0 {
		s = fmt.Sprintf("%2.0f", wholeNum)
	}
	switch fractionNum {
	case 1:
		s += " ⅛"
	case 2:
		s += " ¼"
	case 3:
		s += " ⅜"
	case 4:
		s += " ½"
	case 5:
		s += " ⅝"
	case 6:
		s += " ¾"
	case 7:
		s += " ⅞"
	}
	return
}

func convertCups(cups float64) (amount float64, measure string) {
	amount = cups
	if cups > 0.125 {
		measure = "cup"
	} else if cups > 0.0625 {
		measure = "tablespoon"
		amount *= 16
	} else {
		measure = "teaspoon"
		amount *= 48
	}
	return
}

func FormatMeasure(amount float64, measure string) (s string) {
	if measure == "cup" {
		amount, measure = convertCups(amount)
	}
	s = fmt.Sprintf("%s %s", FormatCookingRational(amount), measure)
	if amount > 1 && measure != "whole" {
		s += "s"
	}
	s = strings.TrimSpace(s)
	return
}

func FormatCost(cost float64) (s string) {
	if cost < 0 {
		s = fmt.Sprintf("Save $%2.2f", math.Abs(cost))
	} else if cost > 0 {
		s = fmt.Sprintf("Lose $%2.2f", math.Abs(cost))
	} else {
		s = "$0"
	}
	return
}

const Year = (365 * 24 * time.Hour)
const Week = (7 * 24 * time.Hour)
const Day = (24 * time.Hour)

func FormatDuration(hours float64) (s string) {
	if hours == 0 {
		return ""
	}
	s = formatDurationRecursively(time.Duration(hours*3600) * time.Second)
	s = strings.TrimSpace(s)
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		s = "no time"
	}
	return
}

func formatDurationRecursively(t time.Duration) (s string) {
	// log.Debug(t)
	if t.Seconds() == 0 {
		return
	}
	timesStrings := []string{"year", "week", "day", "hour", "minute", "second"}
	times := []time.Duration{Year, Week, Day, 1 * time.Hour, 1 * time.Minute, 0 * time.Minute}
	i := 0
	for {
		if t.Seconds() >= times[i].Seconds() {
			break
		}
		i++
	}
	if i == len(times)-1 {
		return
	}
	t2 := int64(t.Seconds() / times[i].Seconds())
	// log.Debug(t.Seconds(), times[i].Seconds())
	t = time.Duration(math.Mod(t.Seconds(), times[i].Seconds())) * time.Second
	if t2 > 1 {
		timesStrings[i] += "s"
	}
	return fmt.Sprintf("%d %s, ", t2, timesStrings[i]) + formatDurationRecursively(t)
}
