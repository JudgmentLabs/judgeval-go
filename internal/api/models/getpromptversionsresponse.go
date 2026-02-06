package models

import (
	"encoding/json"
)

type GetPromptVersionsResponse struct {
	Versions []PromptCommitInfo `json:"versions,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *GetPromptVersionsResponse) UnmarshalJSON(data []byte) error {
	type Alias GetPromptVersionsResponse
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

func (m GetPromptVersionsResponse) MarshalJSON() ([]byte, error) {
	type Alias GetPromptVersionsResponse
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
