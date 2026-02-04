package models

import (
	"encoding/json"
)

type ScorerData struct {
	Id                string      `json:"id,omitempty"`
	Name              string      `json:"name,omitempty"`
	Threshold         float64     `json:"threshold,omitempty"`
	Reason            interface{} `json:"reason,omitempty"`
	EvaluationModel   string      `json:"evaluation_model,omitempty"`
	Error             string      `json:"error,omitempty"`
	ScoreType         string      `json:"score_type,omitempty"`
	NumValue          float64     `json:"num_value,omitempty"`
	BoolValue         bool        `json:"bool_value,omitempty"`
	StrValue          string      `json:"str_value,omitempty"`
	MinimumScoreRange interface{} `json:"minimum_score_range,omitempty"`
	MaximumScoreRange interface{} `json:"maximum_score_range,omitempty"`

	AdditionalProperties map[string]interface{} `json:"-"`
}

func (m *ScorerData) UnmarshalJSON(data []byte) error {
	type Alias ScorerData
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

func (m ScorerData) MarshalJSON() ([]byte, error) {
	type Alias ScorerData
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
