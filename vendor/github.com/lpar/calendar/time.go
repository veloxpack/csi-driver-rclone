package calendar

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

const timeFormat = "15:04:05"

// Time represents a time with no date or timezone information.
// It is compatible with PostgreSQL database TIME values when using the de facto standard lib/pq driver.
type Time time.Time

// NewDate constructs a new Date object for the given year, month and day
func NewTime(h, m, s int) Time {
	return Time(time.Date(0, 0, 0, h, m, s, 0, time.UTC))
}

// UnmarshalJSON unmarshals a Time from JSON format. The date is expected
// to be in RFC 3339 format -- that is, hh:mm:ss in 24 hour clock
func (ti *Time) UnmarshalJSON(b []byte) error {
	sd, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	t, err := time.Parse(timeFormat, sd)
	*ti = Time(t)
	return err
}

// MarshalJSON marshals a Time into JSON format. The date is formatted
// in RFC 3339 format -- that is, hh:mm:ss in 24 hour clock
func (ti Time) MarshalJSON() ([]byte, error) {
	t := time.Time(ti)
	ds := "\"" + t.Format(timeFormat) + "\""
	return []byte(ds), nil
}

// Implement Stringer

// String returns the value of the Time in hh:mm:ss format.
func (ti Time) String() string {
	return time.Time(ti).Format(timeFormat)
}

// Implement Valuer

// Value implements the database/sql Valuer interface.
func (ti Time) Value() (driver.Value, error) {
	return time.Time(ti), nil
}

// Implement Scanner

// Scan implements the database/sql Scanner interface.
func (ti *Time) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("unsupported NULL calendar.Time value")
	}
	t, ok := value.(time.Time)
	if ok {
		*ti = Time(t)
		return nil
	}
	return fmt.Errorf("unable to convert Time")
}

func (ti Time) Equal(other Time) bool {
	tti := time.Time(ti)
	tother := time.Time(other)
	return tti.Second() == tother.Second() &&
		tti.Minute() == tother.Minute() &&
		tti.Second() == tother.Second()
}
