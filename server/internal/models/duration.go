package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Duration is a time.Duration stored as integer nanoseconds. The database
// connection uses the simple query protocol, which sends a plain
// time.Duration as interval text; a bigint column rejects that, so every
// save of a campaign with a batch delay failed.
type Duration time.Duration

func (d Duration) Value() (driver.Value, error) { return int64(d), nil }

func (d *Duration) Scan(v any) error {
	switch n := v.(type) {
	case int64:
		*d = Duration(n)
	case nil:
		*d = 0
	default:
		return fmt.Errorf("cannot scan %T into Duration", v)
	}
	return nil
}
