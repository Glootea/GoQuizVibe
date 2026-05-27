package services_test

import (
	"testing"

	"github.com/goquizvibe/backend/feature/admin/services"
)

func TestPromptGenerator_New(t *testing.T) {
	t.Parallel()

	t.Run("creates prompt generator with nil cache", func(t *testing.T) {
		t.Parallel()
		schema := services.NewQuestionSchema(nil)
		generator := services.NewPromptGenerator(schema)

		if generator == nil {
			t.Fatal("NewPromptGenerator returned nil")
		}
	})
}
