package models

import (
	"encoding/json"
)

type UploadCustomScorerResponse struct {
	ScorerName string `json:"scorer_name,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`

	AdditionalProperties map[string]interface{} `json:"-"`
}

func (m *UploadCustomScorerResponse) UnmarshalJSON(data []byte) error {
	type Alias UploadCustomScorerResponse
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
	m.AdditionalProperties = make(map[string]interface{})
	if err := json.Unmarshal(data, &m.AdditionalProperties); err != nil {
		{
			return err
		}
	}
	return nil
}

func (m UploadCustomScorerResponse) MarshalJSON() ([]byte, error) {
	type Alias UploadCustomScorerResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(&m),
	}

	result := make(map[string]interface{})

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
