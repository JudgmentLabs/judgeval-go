package models

import (
	"encoding/json"
)

type ExperimentRunItem struct {
	OrganizationId  string             `json:"organization_id,omitempty"`
	ExperimentRunId string             `json:"experiment_run_id,omitempty"`
	ExampleId       string             `json:"example_id,omitempty"`
	Data            map[string]any     `json:"data,omitempty"`
	Name            string             `json:"name,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
	Scorers         []ExperimentScorer `json:"scorers,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *ExperimentRunItem) UnmarshalJSON(data []byte) error {
	type Alias ExperimentRunItem
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

func (m ExperimentRunItem) MarshalJSON() ([]byte, error) {
	type Alias ExperimentRunItem
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
