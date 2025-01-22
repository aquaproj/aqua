package vacuum

import "time"

type TimestampChecker struct {
	threshold time.Time
}

func NewTimestampChecker(now time.Time, days int) *TimestampChecker {
	return &TimestampChecker{
		threshold: now.Add(-time.Duration(days) * time.Hour * 24), //nolint:mnd
	}
}

func (c *TimestampChecker) Expired(timestamp time.Time) bool {
	return timestamp.Before(c.threshold)
}
