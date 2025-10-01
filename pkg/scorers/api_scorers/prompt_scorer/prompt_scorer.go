package prompt_scorer

import (
	"fmt"
	"strconv"

	"github.com/JudgmentLabs/judgeval-go/pkg/data"
	"github.com/JudgmentLabs/judgeval-go/pkg/env"
	"github.com/JudgmentLabs/judgeval-go/pkg/internal/api/models"
)

type PromptScorer struct {
	*BasePromptScorer
}

func Get(name string) (*PromptScorer, error) {
	return GetWithCredentials(name, env.JudgmentAPIKey, env.JudgmentOrgID)
}

func GetWithCredentials(name, judgmentAPIKey, organizationID string) (*PromptScorer, error) {
	scorerConfig, err := FetchPromptScorer(name, judgmentAPIKey, organizationID)
	if err != nil {
		return nil, err
	}

	if scorerConfig.IsTrace {
		return nil, fmt.Errorf("scorer with name %s is not a PromptScorer", name)
	}

	options := make(map[string]float64)
	if scorerConfig.Options != nil {
		if optionsMap, ok := scorerConfig.Options.(map[string]interface{}); ok {
			for k, v := range optionsMap {
				if num, ok := v.(float64); ok {
					options[k] = num
				} else if str, ok := v.(string); ok {
					if num, err := strconv.ParseFloat(str, 64); err == nil {
						options[k] = num
					}
				}
			}
		}
	}

	threshold := 0.5
	if scorerConfig.Threshold != 0 {
		threshold = scorerConfig.Threshold
	}

	return &PromptScorer{
		BasePromptScorer: NewBasePromptScorer(
			data.PromptScorer,
			name,
			scorerConfig.Prompt,
			threshold,
			options,
			judgmentAPIKey,
			organizationID,
		),
	}, nil
}

func (ps *PromptScorer) GetScorerConfig() models.ScorerConfig {
	config := ps.BasePromptScorer.APIScorer.GetScorerConfig()

	kwargs := make(map[string]interface{})
	kwargs["prompt"] = ps.GetPrompt()
	if ps.GetOptions() != nil {
		kwargs["options"] = ps.GetOptions()
	}

	if ps.APIScorer.AdditionalProperties != nil {
		for k, v := range ps.APIScorer.AdditionalProperties {
			kwargs[k] = v
		}
	}

	config.Kwargs = kwargs
	return config
}
