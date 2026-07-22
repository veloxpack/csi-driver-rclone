package files_sdk

import "encoding/json"

type Settings struct {
	DesktopDriveMappings map[string]string `json:"desktop_drive_mappings,omitempty" path:"desktop_drive_mappings,omitempty" url:"desktop_drive_mappings,omitempty"`
}

func (s Settings) Identifier() interface{} {
	return nil
}

func (s *Settings) UnmarshalJSON(data []byte) error {
	type settings Settings
	return json.Unmarshal(data, (*settings)(s))
}
