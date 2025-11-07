package v1

import "github.com/JudgmentLabs/judgeval-go/v1/internal/api/models"

type APIScorerType string

const (
	APIScorerTypeFaithfulness         APIScorerType = "faithfulness"
	APIScorerTypeAnswerCorrectness    APIScorerType = "answer_correctness"
	APIScorerTypeAnswerRelevancy      APIScorerType = "answer_relevancy"
	APIScorerTypeInstructionAdherence APIScorerType = "instruction_adherence"
	APIScorerTypeDerailment           APIScorerType = "derailment"
	APIScorerTypePromptScorer         APIScorerType = "prompt_scorer"
	APIScorerTypeTracePromptScorer    APIScorerType = "trace_prompt_scorer"
	APIScorerTypeCustom               APIScorerType = "custom"
)

func (t APIScorerType) String() string {
	return string(t)
}

type ScorerConfig = models.ScorerConfig

type SerializerFunc func(interface{}) (string, error)
