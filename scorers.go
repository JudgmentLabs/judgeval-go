package judgeval

import "github.com/JudgmentLabs/judgeval-go/internal/api"

type ScorersFactory struct {
	client            *api.Client
	projectName       string
	projectID         string
	BuiltIn           *BuiltInScorersFactory
	PromptScorer      *PromptScorerFactory
	TracePromptScorer *PromptScorerFactory
	CustomScorer      *CustomScorerFactory
}

func newScorersFactory(client *api.Client, projectName string, projectID string) *ScorersFactory {
	return &ScorersFactory{
		client:            client,
		projectName:       projectName,
		projectID:         projectID,
		BuiltIn:           &BuiltInScorersFactory{},
		PromptScorer:      &PromptScorerFactory{client: client, projectName: projectName, projectID: projectID, isTrace: false},
		TracePromptScorer: &PromptScorerFactory{client: client, projectName: projectName, projectID: projectID, isTrace: true},
		CustomScorer:      &CustomScorerFactory{},
	}
}
