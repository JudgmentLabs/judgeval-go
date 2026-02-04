package models

import (
	"encoding/json"
)

type TriggerRootSpanRulesResponse struct {
	Success      bool    `json:"success,omitempty"`
	QueuedTraces float64 `json:"queued_traces,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *TriggerRootSpanRulesResponse) UnmarshalJSON(data []byte) error {
	type Alias TriggerRootSpanRulesResponse
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

func (m TriggerRootSpanRulesResponse) MarshalJSON() ([]byte, error) {
	type Alias TriggerRootSpanRulesResponse
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
