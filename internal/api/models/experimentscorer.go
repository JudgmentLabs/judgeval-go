package models

import (
	"encoding/json"
)

type ExperimentScorer struct {
	ScorerDataId       string  `json:"scorer_data_id,omitempty"`
	Name               string  `json:"name,omitempty"`
	Score              float64 `json:"score,omitempty"`
	Success            float64 `json:"success,omitempty"`
	Reason             string  `json:"reason,omitempty"`
	EvaluationModel    string  `json:"evaluation_model,omitempty"`
	Threshold          float64 `json:"threshold,omitempty"`
	CreatedAt          string  `json:"created_at,omitempty"`
	Error              string  `json:"error,omitempty"`
	AdditionalMetadata any     `json:"additional_metadata,omitempty"`
	MinimumScoreRange  float64 `json:"minimum_score_range,omitempty"`
	MaximumScoreRange  float64 `json:"maximum_score_range,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *ExperimentScorer) UnmarshalJSON(data []byte) error {
	type Alias ExperimentScorer
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

func (m ExperimentScorer) MarshalJSON() ([]byte, error) {
	type Alias ExperimentScorer
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
