package models

import (
	"encoding/json"
)

type ExampleEvaluationRun struct {
	Id              string         `json:"id,omitempty"`
	ProjectId       string         `json:"project_id,omitempty"`
	EvalName        string         `json:"eval_name,omitempty"`
	Model           string         `json:"model,omitempty"`
	CreatedAt       string         `json:"created_at,omitempty"`
	UserId          string         `json:"user_id,omitempty"`
	Scorers         []any          `json:"scorers,omitempty"`
	CustomScorers   []BaseScorer   `json:"custom_scorers,omitempty"`
	JudgmentScorers []ScorerConfig `json:"judgment_scorers,omitempty"`
	Examples        []Example      `json:"examples,omitempty"`
	TraceSpanId     string         `json:"trace_span_id,omitempty"`
	TraceId         string         `json:"trace_id,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *ExampleEvaluationRun) UnmarshalJSON(data []byte) error {
	type Alias ExampleEvaluationRun
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

func (m ExampleEvaluationRun) MarshalJSON() ([]byte, error) {
	type Alias ExampleEvaluationRun
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
