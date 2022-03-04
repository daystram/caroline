package util

import (
	"strings"
	"time"
)

func ParseYouTubeDuration(duration string) time.Duration {
	if duration == "" {
		return 0
	}

	duration = strings.Replace(duration, "PT", "", 1)
	duration = strings.Replace(duration, "H", "h", 1)
	duration = strings.Replace(duration, "M", "m", 1)
	duration = strings.Replace(duration, "S", "s", 1)

	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0
	}

	return d
}
