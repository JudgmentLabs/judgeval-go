package models

import (
	"encoding/json"
)

type ExampleScoringResult struct {
	ScorersData    []any   `json:"scorers_data,omitempty"`
	Name           string  `json:"name,omitempty"`
	DataObject     Example `json:"data_object,omitempty"`
	TraceId        string  `json:"trace_id,omitempty"`
	RunDuration    float64 `json:"run_duration,omitempty"`
	EvaluationCost float64 `json:"evaluation_cost,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *ExampleScoringResult) UnmarshalJSON(data []byte) error {
	type Alias ExampleScoringResult
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

func (m ExampleScoringResult) MarshalJSON() ([]byte, error) {
	type Alias ExampleScoringResult
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
