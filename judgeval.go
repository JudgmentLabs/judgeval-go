package judgeval

import (
	"errors"

	"github.com/JudgmentLabs/judgeval-go/env"
	"github.com/JudgmentLabs/judgeval-go/internal/api"
)

type Judgeval struct {
	apiClient   *api.Client
	projectName string
	projectID   string
	Tracer      *TracerFactory
	Scorers     *ScorersFactory
	Evaluation  *EvaluationFactory
}

func NewJudgeval(projectName string, opts ...Option) (*Judgeval, error) {
	cfg := &clientConfig{
		apiKey: env.JudgmentAPIKey,
		orgID:  env.JudgmentOrgID,
		apiURL: env.JudgmentAPIURL,
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	if cfg.apiKey == "" {
		return nil, errors.New("API key is required: set JUDGMENT_API_KEY environment variable or use WithAPIKey option")
	}
	if cfg.orgID == "" {
		return nil, errors.New("organization ID is required: set JUDGMENT_ORG_ID environment variable or use WithOrganizationID option")
	}
	if cfg.apiURL == "" {
		return nil, errors.New("API URL is required: set JUDGMENT_API_URL environment variable or use WithAPIURL option")
	}
	if projectName == "" {
		return nil, errors.New("project name is required")
	}

	apiClient := api.NewClient(cfg.apiURL, cfg.apiKey, cfg.orgID)
	projectID, err := resolveProjectID(apiClient, projectName)
	if err != nil {
		return nil, err
	}

	return &Judgeval{
		apiClient:   apiClient,
		projectName: projectName,
		projectID:   projectID,
		Tracer:      &TracerFactory{client: apiClient, projectName: projectName, projectID: projectID},
		Scorers:     newScorersFactory(apiClient, projectName, projectID),
		Evaluation:  &EvaluationFactory{client: apiClient, projectName: projectName, projectID: projectID},
	}, nil
}
