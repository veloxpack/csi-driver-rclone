package calendar

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type NullDate struct {
	Date  Date
	Valid bool
}

// NewNullDate constructs a new NullDate object for the given year, month and day
func NewNullDate(y, m, d int) NullDate {
	return NullDate{
		Date:  NewDate(y, m, d),
		Valid: true,
	}
}

// NullDateFromTime constructs a new NullDate object from the provided time.Time value, throwing away all time and timezone information.
func NullDateFromTime(t time.Time) NullDate {
	return NewNullDate(t.Year(), int(t.Month()), t.Day())
}

// UnmarshalJSON unmarshals a NullDate from JSON format. The date is expected
// to be in full-date format as per RFC 3339 -- that is, yyyy-mm-dd.
func (d *NullDate) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == "" {
		d.Valid = false
		return nil
	}
	sd, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	t, err := time.Parse(dateFormat, sd)
	*d = NullDate{
		Date:  Date(t),
		Valid: true,
	}
	return err
}

// MarshalJSON marshals a NullDate into JSON format. The date is formatted
// in RFC 3339 full-date format -- that is, yyyy-mm-dd.
func (d NullDate) MarshalJSON() ([]byte, error) {
	var ds string
	if d.Valid {
		t := time.Time(d.Date)
		ds = "\"" + t.Format(dateFormat) + "\""
	} else {
		ds = "null"
	}
	return []byte(ds), nil
}

// Implement Stringer

// String returns the value of the NullDate in ISO-8601 / RFC 3339 format yyyy-mm-dd.
func (d NullDate) String() string {
	if d.Valid {
		return d.Date.String()
	}
	return "null"
}

// Implement Valuer

// Value implements the database/sql Valuer interface.
func (d NullDate) Value() (driver.Value, error) {
	if d.Valid {
		return time.Time(d.Date), nil
	}
	return nil, nil
}

// Implement Scanner

// Scan implements the database/sql Scanner interface.
func (d *NullDate) Scan(value interface{}) error {
	if value == nil {
		d.Valid = false
		return nil
	}
	t, ok := value.(time.Time)
	if ok {
		d.Date = DateFromTime(t)
		d.Valid = true
		return nil
	}
	return fmt.Errorf("unable to convert NullDate")
}

// Equal returns true if the two dates are equal.
func (d NullDate) Equal(other NullDate) bool {
	if !d.Valid && !other.Valid {
		return true // both null means equal
	}
	if d.Valid != other.Valid {
		return false // different nullness means not equal
	}
	return d.Date.Equal(other.Date)
}
