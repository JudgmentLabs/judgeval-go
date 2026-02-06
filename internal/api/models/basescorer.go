package models

import (
	"encoding/json"
)

type BaseScorer struct {
	ScoreType          string         `json:"score_type,omitempty"`
	Name               string         `json:"name,omitempty"`
	ClassName          string         `json:"class_name,omitempty"`
	Score              float64        `json:"score,omitempty"`
	MinimumScoreRange  float64        `json:"minimum_score_range,omitempty"`
	MaximumScoreRange  float64        `json:"maximum_score_range,omitempty"`
	ScoreBreakdown     map[string]any `json:"score_breakdown,omitempty"`
	Reason             any            `json:"reason,omitempty"`
	Success            bool           `json:"success,omitempty"`
	Model              string         `json:"model,omitempty"`
	Error              string         `json:"error,omitempty"`
	AdditionalMetadata map[string]any `json:"additional_metadata,omitempty"`
	User               string         `json:"user,omitempty"`
	ServerHosted       bool           `json:"server_hosted,omitempty"`
	UsingNativeModel   bool           `json:"using_native_model,omitempty"`
	RequiredParams     []string       `json:"required_params,omitempty"`
	StrictMode         bool           `json:"strict_mode,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *BaseScorer) UnmarshalJSON(data []byte) error {
	type Alias BaseScorer
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

func (m BaseScorer) MarshalJSON() ([]byte, error) {
	type Alias BaseScorer
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
