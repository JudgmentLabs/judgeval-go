package judgeval

import (
	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
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

func (f *CustomScorerFactory) Get(name string, className string) (*CustomScorer, error) {
	return &CustomScorer{
		name:         name,
		className:    className,
		serverHosted: true,
	}, nil
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

func (s *CustomScorer) GetBaseScorer() models.BaseScorer {
	return models.BaseScorer{
		ScoreType: APIScorerTypeCustom.String(),
		Name:      s.name,
	}
}
