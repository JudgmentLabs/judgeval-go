package models

import (
	"encoding/json"
)

type PromptCommitInfo struct {
	Name           string   `json:"name,omitempty"`
	Prompt         string   `json:"prompt,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	CommitId       string   `json:"commit_id,omitempty"`
	ParentCommitId string   `json:"parent_commit_id,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	FirstName      string   `json:"first_name,omitempty"`
	LastName       string   `json:"last_name,omitempty"`
	UserEmail      string   `json:"user_email,omitempty"`

	AdditionalProperties map[string]any `json:"-"`
}

func (m *PromptCommitInfo) UnmarshalJSON(data []byte) error {
	type Alias PromptCommitInfo
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

func (m PromptCommitInfo) MarshalJSON() ([]byte, error) {
	type Alias PromptCommitInfo
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
