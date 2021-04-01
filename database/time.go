package database

import (
	"fmt"
	"time"
)

type TimeGroup interface {
}

func (it *Db) ParseTime(t string) (time.Time, error) {
	switch it.Type {
	case "mysql":
		return time.Parse("2006-01-02T15:04:05Z", t)
	case "postgres":
		return time.Parse("2006-01-02T15:04:05Z", t)
	default:
		return time.Parse("2006-01-02 15:04:05", t)
	}
}

// FormatTime returns the timestamp in the same format as the DATETIME column in database
func (it *Db) FormatTime(t time.Time) string {
	switch it.Type {
	case "postgres":
		return t.Format("2006-01-02 15:04:05.999999999")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}

// SelectByTime returns an SQL query that will group "created_at" column by x seconds and returns as "timeframe"
func (it *Db) SelectByTime(increment time.Duration) string {
	seconds := int64(increment.Seconds())
	switch it.Type {
	case "mysql":
		return fmt.Sprintf("FROM_UNIXTIME(FLOOR(UNIX_TIMESTAMP(created_at) / %d) * %d) AS timeframe", seconds, seconds)
	case "postgres":
		return fmt.Sprintf("date_trunc('minute', created_at) - (CAST(EXTRACT(MINUTE FROM created_at) AS integer) %% %d) * interval '1 minute' AS timeframe", int64(increment.Minutes()))
	default:
		return fmt.Sprintf("datetime((strftime('%%s', created_at) / %d) * %d, 'unixepoch') as timeframe", seconds, seconds)
	}
}
