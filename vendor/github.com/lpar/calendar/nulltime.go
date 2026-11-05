package calendar

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type NullTime struct {
	Time  Time
	Valid bool
}

// NewNullTime constructs a new NullTime object for the given hours, minutes and seconds
func NewNullTime(h, m, s int) NullTime {
	return NullTime{
		Time:  NewTime(h, m, s),
		Valid: true,
	}
}

// NullTimeFromTime constructs a new NullTime object from the provided time.Time value, throwing away all time and timezone information.
func NullTimeFromTime(t time.Time) NullTime {
	return NewNullTime(t.Hour(), t.Minute(), t.Second())
}

// UnmarshalJSON unmarshals a NullTime from JSON format. The date is expected
// to be in full-date format as per RFC 3339 -- that is, yyyy-mm-dd.
func (t *NullTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == "" {
		t.Valid = false
		return nil
	}
	st, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	ti, err := time.Parse(timeFormat, st)
	*t = NullTime{
		Time:  Time(ti),
		Valid: true,
	}
	return err
}

// MarshalJSON marshals a NullTime into JSON format. The date is formatted
// in RFC 3339 full-date format -- that is, yyyy-mm-dd.
func (t NullTime) MarshalJSON() ([]byte, error) {
	var ds string
	if t.Valid {
		ti := time.Time(t.Time)
		ds = "\"" + ti.Format(timeFormat) + "\""
	} else {
		ds = "null"
	}
	return []byte(ds), nil
}

// Implement Stringer

// String returns the value of the NullTime in ISO-8601 / RFC 3339 format yyyy-mm-dd.
func (t NullTime) String() string {
	if t.Valid {
		return t.Time.String()
	}
	return "null"
}

// Implement Valuer

// Value implements the database/sql Valuer interface.
func (t NullTime) Value() (driver.Value, error) {
	if t.Valid {
		return time.Time(t.Time), nil
	}
	return nil, nil
}

// Implement Scanner

// Scan implements the database/sql Scanner interface.
func (t *NullTime) Scan(value interface{}) error {
	if value == nil {
		t.Valid = false
		return nil
	}
	ti, ok := value.(time.Time)
	if ok {
		t.Time = Time(ti)
		t.Valid = true
		return nil
	}
	return fmt.Errorf("unable to convert NullTime")
}

// Equal returns true if the two dates are equal.
func (t NullTime) Equal(other NullTime) bool {
	if !t.Valid && !other.Valid {
		return true // both null means equal
	}
	if t.Valid != other.Valid {
		return false // different nullness means not equal
	}
	return t.Time.Equal(other.Time)
}
