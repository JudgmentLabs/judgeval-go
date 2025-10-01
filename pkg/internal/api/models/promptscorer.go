package models

import (
	"encoding/json"
)

type PromptScorer struct {
	Name        string      `json:"name,omitempty"`
	Prompt      string      `json:"prompt,omitempty"`
	Threshold   float64     `json:"threshold,omitempty"`
	Model       string      `json:"model,omitempty"`
	Options     interface{} `json:"options,omitempty"`
	Description string      `json:"description,omitempty"`
	CreatedAt   string      `json:"created_at,omitempty"`
	UpdatedAt   string      `json:"updated_at,omitempty"`
	IsTrace     bool        `json:"is_trace,omitempty"`

	AdditionalProperties map[string]interface{} `json:"-"`
}

func (m *PromptScorer) UnmarshalJSON(data []byte) error {
	type Alias PromptScorer
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

func (m PromptScorer) MarshalJSON() ([]byte, error) {
	type Alias PromptScorer
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
