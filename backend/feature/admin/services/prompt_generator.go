package services

import (
	"context"
	"fmt"

	"github.com/goquizvibe/backend/shared/locales"
)

type PromptGenerator struct {
	schema *QuestionSchema
}

func NewPromptGenerator(schema *QuestionSchema) *PromptGenerator {
	return &PromptGenerator{schema: schema}
}

func (g *PromptGenerator) GetSchema(ctx context.Context) (string, error) {
	return g.schema.GetSchema(ctx)
}

func (g *PromptGenerator) GeneratePrompt(ctx context.Context, quizTitle string, tr locales.Translator) (string, error) {
	schemaJSON, err := g.schema.GetSchema(ctx)
	if err != nil {
		return "", fmt.Errorf("get schema: %w", err)
	}

	prompt := tr.YouAreAnExpertAtCreatingTestQuestionsCreateAnArrayOfQuestionsOnTheTopic() + "\n\n" +
		tr.Topic() + ": " + quizTitle + "\n\n" +
		tr.ResponseMustBeAValidJsonArrayOfObjectsNoAdditionalText() + "\n\n" +
		tr.JsonSchemaForQuestionObject() + "\n" + schemaJSON

	return prompt, nil
}