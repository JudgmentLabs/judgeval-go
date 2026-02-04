package models

import (
	"encoding/json"
)

type TraceEvaluationRun struct {
	Id              string          `json:"id,omitempty"`
	ProjectId       string          `json:"project_id,omitempty"`
	EvalName        string          `json:"eval_name,omitempty"`
	Model           string          `json:"model,omitempty"`
	CreatedAt       string          `json:"created_at,omitempty"`
	UserId          string          `json:"user_id,omitempty"`
	Scorers         []interface{}   `json:"scorers,omitempty"`
	CustomScorers   []BaseScorer    `json:"custom_scorers,omitempty"`
	JudgmentScorers []ScorerConfig  `json:"judgment_scorers,omitempty"`
	TraceAndSpanIds [][]interface{} `json:"trace_and_span_ids,omitempty"`
	IsOffline       bool            `json:"is_offline,omitempty"`
	IsBehavior      bool            `json:"is_behavior,omitempty"`

	AdditionalProperties map[string]interface{} `json:"-"`
}

func (m *TraceEvaluationRun) UnmarshalJSON(data []byte) error {
	type Alias TraceEvaluationRun
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

func (m TraceEvaluationRun) MarshalJSON() ([]byte, error) {
	type Alias TraceEvaluationRun
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
