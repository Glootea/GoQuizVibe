package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/invopop/jsonschema"
	cache "github.com/goquizvibe/backend/shared/infrastructure/cache"
)

const (
	questionSchemaCacheKey = "question_schema"
	questionSchemaTTL      = 15 * time.Minute
)

type QuestionSchema struct {
	cache *cache.CacheService
}

func NewQuestionSchema(cache *cache.CacheService) *QuestionSchema {
	return &QuestionSchema{cache: cache}
}

func (s *QuestionSchema) GetSchema(ctx context.Context) (string, error) {
	var schemaJSON string
	if s.cache.Get(ctx, questionSchemaCacheKey, &schemaJSON, "question_schema") {
		return schemaJSON, nil
	}

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
	}
	if err := reflector.AddGoComments("github.com/goquizvibe/backend/feature/admin/services", "question_schema.go"); err != nil {
		return "", err
	}
	schema := reflector.Reflect(&QuestionImportSchema{})

	data, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	schemaJSON = string(data)

	_ = s.cache.Set(ctx, questionSchemaCacheKey, schemaJSON, questionSchemaTTL, "question_schema")

	return schemaJSON, nil
}

type QuestionImportSchema struct {
	// question text
	Text string `json:"text"`
	// one of: choice, open, fill
	Type string `json:"type" jsonschema:"enum=choice,enum=open,enum=fill"`
	// answer options for choice type, empty for others
	Options []string `json:"options"`
	// correct answer
	CorrectAnswer string `json:"correct_answer"`
	// explanation for the answer
	Explanation string `json:"explanation"`
	// point value for the question (default 10)
	Points int `json:"points" jsonschema:"minimum=1"`
	// position in the quiz (0-based)
	OrderIndex int `json:"order_index" jsonschema:"minimum=0"`
}
