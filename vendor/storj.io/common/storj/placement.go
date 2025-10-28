// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"database/sql/driver"
	"strconv"

	"github.com/zeebo/errs"
)

// PlacementConstraint is the ID of the placement/geofencing rule.
type PlacementConstraint uint16

const (
	// DefaultPlacement placement is used, when no specific placement rule is defined.
	DefaultPlacement PlacementConstraint = 0

	// EveryCountry includes all countries.
	// Deprecated: use DefaultPlacement, which may exclude some nodes based on placement configuration.
	EveryCountry PlacementConstraint = 0

	// EU includes only the 27 members of European Union.
	// Deprecated: placement definitions depend on the configuration.
	EU PlacementConstraint = 1

	// EEA defines the European Economic Area (EU + 3 countries), the area where GDPR is valid.
	// Deprecated: placement definitions depend on the configuration.
	EEA PlacementConstraint = 2

	// US filters nodes only from the United States.
	// Deprecated: placement definitions depend on the configuration.
	US PlacementConstraint = 3

	// DE placement uses nodes only from Germany.
	// Deprecated: placement definitions depend on the configuration.
	DE PlacementConstraint = 4

	// InvalidPlacement is used when there is no information about the stored placement.
	// Deprecated: placement definitions depend on the configuration.
	InvalidPlacement PlacementConstraint = 5

	// NR placement uses nodes that are not in RU or other countries sanctioned because of the RU/UA War.
	// Deprecated: placement definitions depend on the configuration.
	NR PlacementConstraint = 6
)

// Value implements the driver.Valuer interface.
func (p PlacementConstraint) Value() (driver.Value, error) {
	return int64(p), nil
}

// Scan implements the sql.Scanner interface.
func (p *PlacementConstraint) Scan(value any) error {
	if value == nil {
		*p = DefaultPlacement
		return nil
	}

	if _, isInt64 := value.(int64); !isInt64 {
		return errs.New("unable to scan %T into PlacementConstraint", value)
	}

	code, err := driver.Int32.ConvertValue(value)
	if err != nil {
		return errs.Wrap(err)
	}
	*p = PlacementConstraint(uint16(code.(int64)))
	return nil
}

// EncodeSpanner implements spanner.Encoder.
func (p PlacementConstraint) EncodeSpanner() (any, error) {
	return p.Value()
}

// DecodeSpanner implements spanner.Decoder.
func (p *PlacementConstraint) DecodeSpanner(input any) error {
	var sVal string
	switch val := input.(type) {
	case string:
		sVal = val
	case *string:
		if val == nil {
			*p = DefaultPlacement
			return nil
		}
		sVal = *val
	default:
		return errs.New("unable to decode %q to PlacementConstraint", input)
	}
	iVal, err := strconv.ParseInt(sVal, 10, 64)
	if err != nil {
		return err
	}
	if int64(PlacementConstraint(iVal)) != iVal {
		return errs.New("value out of bounds for PlacementConstraint: %d", iVal)
	}
	*p = PlacementConstraint(iVal)
	return nil
}
