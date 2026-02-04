package models

import (
	"encoding/json"
)

type AddToRunEvalQueueTracesResponse struct {
	Success bool   `json:"success,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`

	AdditionalProperties map[string]interface{} `json:"-"`
}

func (m *AddToRunEvalQueueTracesResponse) UnmarshalJSON(data []byte) error {
	type Alias AddToRunEvalQueueTracesResponse
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

func (m AddToRunEvalQueueTracesResponse) MarshalJSON() ([]byte, error) {
	type Alias AddToRunEvalQueueTracesResponse
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
