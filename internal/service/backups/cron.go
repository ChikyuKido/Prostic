package backups

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func cronMatches(expression string, now time.Time) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(expression))
	if len(fields) != 5 {
		return false, fmt.Errorf("invalid cron expression")
	}

	minute, err := matchCronField(fields[0], now.Minute(), 0, 59, false)
	if err != nil || !minute {
		return false, err
	}
	hour, err := matchCronField(fields[1], now.Hour(), 0, 23, false)
	if err != nil || !hour {
		return false, err
	}
	month, err := matchCronField(fields[3], int(now.Month()), 1, 12, false)
	if err != nil || !month {
		return false, err
	}

	domWildcard := fields[2] == "*"
	dowWildcard := fields[4] == "*"
	dom, err := matchCronField(fields[2], now.Day(), 1, 31, false)
	if err != nil {
		return false, err
	}
	dow, err := matchCronField(fields[4], int(now.Weekday()), 0, 7, true)
	if err != nil {
		return false, err
	}

	if domWildcard && dowWildcard {
		return true, nil
	}
	if domWildcard {
		return dow, nil
	}
	if dowWildcard {
		return dom, nil
	}

	return dom || dow, nil
}

func matchCronField(field string, value int, min int, max int, sundayWrap bool) (bool, error) {
	parts := strings.Split(field, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false, fmt.Errorf("invalid cron field")
		}

		rangePart := part
		step := 1
		if strings.Contains(part, "/") {
			split := strings.SplitN(part, "/", 2)
			rangePart = split[0]
			parsedStep, err := strconv.Atoi(split[1])
			if err != nil || parsedStep <= 0 {
				return false, fmt.Errorf("invalid cron step")
			}
			step = parsedStep
		}

		start := min
		end := max
		if rangePart != "*" {
			if strings.Contains(rangePart, "-") {
				split := strings.SplitN(rangePart, "-", 2)
				parsedStart, err := strconv.Atoi(split[0])
				if err != nil {
					return false, err
				}
				parsedEnd, err := strconv.Atoi(split[1])
				if err != nil {
					return false, err
				}
				start = parsedStart
				end = parsedEnd
			} else {
				parsedValue, err := strconv.Atoi(rangePart)
				if err != nil {
					return false, err
				}
				start = parsedValue
				end = parsedValue
			}
		}

		if sundayWrap {
			if start == 7 {
				start = 0
			}
			if end == 7 {
				end = 0
			}
		}

		if rangePart == "*" {
			start = min
			end = max
		}

		if start < min || end > max || end < start {
			return false, fmt.Errorf("cron value out of range")
		}

		for candidate := start; candidate <= end; candidate += step {
			cmp := candidate
			if sundayWrap && cmp == 7 {
				cmp = 0
			}
			if cmp == value {
				return true, nil
			}
		}
	}

	return false, nil
}
