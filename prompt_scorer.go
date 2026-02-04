package judgeval

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"sync"

	"github.com/JudgmentLabs/judgeval-go/env"
	"github.com/JudgmentLabs/judgeval-go/internal/api"
	"github.com/JudgmentLabs/judgeval-go/internal/api/models"
)

type PromptScorerFactory struct {
	client      *api.Client
	projectName string
	projectID   string
	isTrace     bool
	cache       sync.Map
}

type PromptScorer struct {
	name        string
	prompt      string
	threshold   float64
	options     map[string]float64
	model       string
	description string
	isTrace     bool
}

func (f *PromptScorerFactory) Get(ctx context.Context, name string) (*PromptScorer, error) {
	cacheKey := f.buildCacheKey(name)

	if cached, ok := f.cache.Load(cacheKey); ok {
		return cached.(*PromptScorer), nil
	}

	names := name
	isTrace := strconv.FormatBool(f.isTrace)
	resp, err := f.client.GetProjectsScorers(f.projectID, &names, &isTrace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prompt scorer '%s': %w", name, err)
	}

	if len(resp.Scorers) == 0 {
		return nil, fmt.Errorf("failed to fetch prompt scorer '%s': not found", name)
	}

	scorerModel := resp.Scorers[0]

	if scorerModel.IsTrace != f.isTrace {
		expectedType := "PromptScorer"
		actualType := "PromptScorer"
		if f.isTrace {
			expectedType = "TracePromptScorer"
		}
		if scorerModel.IsTrace {
			actualType = "TracePromptScorer"
		}
		return nil, fmt.Errorf("scorer with name %s is a %s, not a %s", name, actualType, expectedType)
	}

	options := make(map[string]float64)
	if scorerModel.Options != nil {
		if optsMap, ok := scorerModel.Options.(map[string]any); ok {
			for k, v := range optsMap {
				if floatVal, ok := v.(float64); ok {
					options[k] = floatVal
				}
			}
		}
	}

	threshold := 0.5
	if scorerModel.Threshold != 0 {
		threshold = scorerModel.Threshold
	}

	modelName := env.JudgmentDefaultGPTModel
	if scorerModel.Model != "" {
		modelName = scorerModel.Model
	}

	description := ""
	if scorerModel.Description != "" {
		description = scorerModel.Description
	}

	scorer := &PromptScorer{
		name:        name,
		prompt:      scorerModel.Prompt,
		threshold:   threshold,
		options:     options,
		model:       modelName,
		description: description,
		isTrace:     f.isTrace,
	}

	f.cache.Store(cacheKey, scorer)
	return scorer, nil
}

func (f *PromptScorerFactory) buildCacheKey(name string) string {
	return fmt.Sprintf("%s:%s:%s:%s", f.projectID, name, f.client.GetAPIKey(), f.client.GetOrganizationID())
}

func (s *PromptScorer) GetName() string {
	return s.name
}

func (s *PromptScorer) GetPrompt() string {
	return s.prompt
}

func (s *PromptScorer) GetThreshold() float64 {
	return s.threshold
}

func (s *PromptScorer) GetOptions() map[string]float64 {
	optsCopy := make(map[string]float64)
	maps.Copy(optsCopy, s.options)
	return optsCopy
}

func (s *PromptScorer) GetModel() string {
	return s.model
}

func (s *PromptScorer) GetDescription() string {
	return s.description
}

func (s *PromptScorer) SetThreshold(threshold float64) {
	s.threshold = threshold
}

func (s *PromptScorer) SetPrompt(prompt string) {
	s.prompt = prompt
}

func (s *PromptScorer) SetModel(model string) {
	s.model = model
}

func (s *PromptScorer) SetOptions(options map[string]float64) {
	s.options = make(map[string]float64)
	maps.Copy(s.options, options)
}

func (s *PromptScorer) SetDescription(description string) {
	s.description = description
}

func (s *PromptScorer) AppendToPrompt(addition string) {
	s.prompt = s.prompt + addition
}

func (s *PromptScorer) GetScorerConfig() *models.ScorerConfig {
	scoreType := APIScorerTypePromptScorer.String()
	if s.isTrace {
		scoreType = APIScorerTypeTracePromptScorer.String()
	}

	kwargs := map[string]interface{}{
		"prompt": s.prompt,
	}

	if len(s.options) > 0 {
		kwargs["options"] = s.options
	}
	if s.model != "" {
		kwargs["model"] = s.model
	}
	if s.description != "" {
		kwargs["description"] = s.description
	}

	return &models.ScorerConfig{
		ScoreType: scoreType,
		Threshold: s.threshold,
		Name:      s.name,
		Kwargs:    kwargs,
	}
}
