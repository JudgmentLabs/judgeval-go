package models

import (
	"encoding/json"
)

type UploadCustomScorerRequest struct {
	ScorerName       string  `json:"scorer_name,omitempty"`
	ScorerCode       string  `json:"scorer_code,omitempty"`
	RequirementsText string  `json:"requirements_text,omitempty"`
	ClassName        string  `json:"class_name,omitempty"`
	Overwrite        bool    `json:"overwrite,omitempty"`
	ScorerType       string  `json:"scorer_type,omitempty"`
	ResponseType     string  `json:"response_type,omitempty"`
	Version          float64 `json:"version,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *UploadCustomScorerRequest) UnmarshalJSON(data []byte) error {
	type Alias UploadCustomScorerRequest
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

func (m UploadCustomScorerRequest) MarshalJSON() ([]byte, error) {
	type Alias UploadCustomScorerRequest
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
