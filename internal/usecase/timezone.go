package usecase

import (
	"time"
)

// getGMT7Location returns Asia/Jakarta timezone (GMT+7)
func getGMT7Location() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to UTC+7 if Asia/Jakarta is not available
		return time.FixedZone("GMT+7", 7*60*60)
	}
	return loc
}

// nowGMT7 returns current time in GMT+7 timezone
func nowGMT7() time.Time {
	return time.Now().In(getGMT7Location())
}

// dateGMT7 creates a time.Time with GMT+7 timezone
func dateGMT7(year int, month time.Month, day, hour, min, sec, nsec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, nsec, getGMT7Location())
}
