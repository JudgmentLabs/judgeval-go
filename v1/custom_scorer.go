package v1

import (
	"fmt"

	"github.com/JudgmentLabs/judgeval-go/v1/internal/api/models"
)

type CustomScorerFactory struct{}

type CustomScorerParams struct {
	Name      string
	ClassName *string
}

type CustomScorer struct {
	name         string
	className    string
	serverHosted bool
}

func (f *CustomScorerFactory) Get(name string, className *string) *CustomScorer {
	finalClassName := name
	if className != nil {
		finalClassName = *className
	}

	return &CustomScorer{
		name:         name,
		className:    finalClassName,
		serverHosted: true,
	}
}

func (f *CustomScorerFactory) Create(params CustomScorerParams) (*CustomScorer, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return f.Get(params.Name, params.ClassName), nil
}

func (s *CustomScorer) GetName() string {
	return s.name
}

func (s *CustomScorer) GetClassName() string {
	return s.className
}

func (s *CustomScorer) IsServerHosted() bool {
	return s.serverHosted
}

func (s *CustomScorer) GetScorerConfig() *models.ScorerConfig {
	return &models.ScorerConfig{
		ScoreType: APIScorerTypeCustom.String(),
		Name:      s.name,
		Kwargs: map[string]interface{}{
			"class_name":    s.className,
			"server_hosted": s.serverHosted,
		},
	}
}
