package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type HolidayRegion struct {
	Code string `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	Name string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
}

// Identifier no path or id

type HolidayRegionCollection []HolidayRegion

type HolidayRegionGetSupportedParams struct {
	ListParams
}

func (h *HolidayRegion) UnmarshalJSON(data []byte) error {
	type holidayRegion HolidayRegion
	var v holidayRegion
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*h = HolidayRegion(v)
	return nil
}

func (h *HolidayRegionCollection) UnmarshalJSON(data []byte) error {
	type holidayRegions HolidayRegionCollection
	var v holidayRegions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*h = HolidayRegionCollection(v)
	return nil
}

func (h *HolidayRegionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*h))
	for i, v := range *h {
		ret[i] = v
	}

	return &ret
}
