package api_scorers

import (
	"github.com/JudgmentLabs/judgeval-go/src/data"
	"github.com/JudgmentLabs/judgeval-go/src/scorers"
)

type AnswerCorrectnessScorer struct {
	*scorers.APIScorer
}

func NewAnswerCorrectnessScorer(options ...scorers.APIScorerOption) *AnswerCorrectnessScorer {

	allOptions := append(options, scorers.WithRequiredParams([]string{
		"input",
		"actual_output",
		"expected_output",
	}))

	apiScorer := scorers.NewAPIScorer(data.AnswerCorrectness, allOptions...)

	return &AnswerCorrectnessScorer{
		APIScorer: apiScorer,
	}
}
