package models

import (
	"encoding/json"
)

type FetchExperimentRunResponse struct {
	Results      []ExperimentRunItem `json:"results,omitempty"`
	UiResultsUrl string              `json:"ui_results_url,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *FetchExperimentRunResponse) UnmarshalJSON(data []byte) error {
	type Alias FetchExperimentRunResponse
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

func (m FetchExperimentRunResponse) MarshalJSON() ([]byte, error) {
	type Alias FetchExperimentRunResponse
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
