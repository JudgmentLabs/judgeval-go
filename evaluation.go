package judgeval

import "github.com/JudgmentLabs/judgeval-go/internal/api"

type EvaluationFactory struct {
	client      *api.Client
	projectName string
	projectID   string
}

type EvaluationCreateParams struct {
}

type Evaluation struct {
	client      *api.Client
	projectName string
	projectID   string
}

func (f *EvaluationFactory) Create(params EvaluationCreateParams) *Evaluation {
	return &Evaluation{
		client:      f.client,
		projectName: f.projectName,
		projectID:   f.projectID,
	}
}
