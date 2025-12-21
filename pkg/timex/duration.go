package timex

import (
	"strconv"
	"time"
)

// Duration support days with constant 24 hour duration
type Duration string

func (l Duration) Duration() time.Duration {
	raw := string(l)

	duration, err := time.ParseDuration(raw)
	if err == nil {
		return duration
	}

	// support "Xd" syntax
	if len(raw) > 1 && raw[len(raw)-1] == 'd' {
		days, err := strconv.Atoi(raw[:len(raw)-1])
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	return 0
}

func IsValidDuration(duration string) bool {
	_, err := time.ParseDuration(duration)
	if err != nil {
		isD := duration[len(duration)-1] == 'd'
		if isD {
			rest := duration[:len(duration)-1]
			_, err := strconv.Atoi(rest)
			return err == nil
		}
	}
	return true
}
