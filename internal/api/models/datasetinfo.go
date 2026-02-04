package models

import (
	"encoding/json"
)

type DatasetInfo struct {
	DatasetId string  `json:"dataset_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	Kind      string  `json:"kind,omitempty"`
	Entries   float64 `json:"entries,omitempty"`
	Creator   string  `json:"creator,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *DatasetInfo) UnmarshalJSON(data []byte) error {
	type Alias DatasetInfo
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		{
			return err
		}
	}
	m.AdditionalProperties = make(map[string]any)
	if err := json.Unmarshal(data, &m.AdditionalProperties); err != nil {
		{
			return err
		}
	}
	return nil
}

func (m DatasetInfo) MarshalJSON() ([]byte, error) {
	type Alias DatasetInfo
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(&m),
	}

	result := make(map[string]any)

	mainBytes, err := json.Marshal(aux)
	if err != nil {
		{
			return nil, err
		}
	}

	if err := json.Unmarshal(mainBytes, &result); err != nil {
		{
			return nil, err
		}
	}

	for k, v := range m.AdditionalProperties {
		{
			result[k] = v
		}
	}

	return json.Marshal(result)
}
