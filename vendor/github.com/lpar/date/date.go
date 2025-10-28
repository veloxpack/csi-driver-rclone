package date

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// Date represents a date with no time or timezone information.
// It is compatible with PostgreSQL database DATE values when using the de facto standard lib/pq driver.
type Date time.Time

// NewDate constructs a new Date object for the given year, month and day
func NewDate(y, m, d int) Date {
	return Date(time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC))
}

// FromTime constructs a new Date object from the provided time.Time value, throwing away all time and timezone information.
func FromTime(t time.Time) Date {
	return NewDate(t.Year(), int(t.Month()), t.Day())
}

// UnmarshalJSON unmarshals a Date from JSON format. The date is expected
// to be in full-date format as per RFC 3339 -- that is, yyyy-mm-dd.
func (d *Date) UnmarshalJSON(b []byte) error {
	sd, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", sd)
	*d = Date(t)
	return err
}

// MarshalJSON marshals a Date into JSON format. The date is formatted
// in RFC 3339 full-date format -- that is, yyyy-mm-dd.
func (d *Date) MarshalJSON() ([]byte, error) {
	t := time.Time(*d)
	ds := "\"" + t.Format("2006-01-02") + "\""
	return []byte(ds), nil
}

// Implement Stringer

// String returns the value of the Date in ISO-8601 / RFC 3339 format yyyy-mm-dd.
func (d *Date) String() string {
	return time.Time(*d).Format("2006-01-02")
}

// Implement Valuer

// Value implements the database/sql Valuer interface.
func (d Date) Value() (driver.Value, error) {
	return time.Time(d), nil
}

// Implement Scanner

// Scan implements the database/sql Scanner interface.
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("unsupported NULL date.Date value")
	}
	t, ok := value.(time.Time)
	if ok {
		*d = Date(t)
		return nil
	}
	return fmt.Errorf("unable to convert Date")
}

// Before returns true if the first date (the reciever) is before the second date (the argument).
func (d *Date) Before(other Date) bool {
	return time.Time(*d).Before(time.Time(other))
}

// After returns true if the first date (the reciever) is after the second date (the argument).
func (d *Date) After(other Date) bool {
	return time.Time(*d).After(time.Time(other))
}

// Equal returns true if the two dates are equal.
func (d *Date) Equal(other Date) bool {
	return time.Time(*d).Equal(time.Time(other))
}

// AddDate adds the specified number of years, months and days to the Date, returning another Date.
func (d *Date) AddDate(yy int, mm int, dd int) Date {
	return Date(time.Time(*d).AddDate(yy, mm, dd))
}
