package services_test

import (
	"encoding/json"
	"testing"

	"github.com/goquizvibe/backend/feature/admin/services"
)

func TestQuestionSchema_New(t *testing.T) {
	t.Parallel()

	t.Run("creates schema with nil cache", func(t *testing.T) {
		t.Parallel()
		schema := services.NewQuestionSchema(nil)

		if schema == nil {
			t.Fatal("NewQuestionSchema returned nil")
		}
	})
}

func TestQuestionImportSchema(t *testing.T) {
	t.Parallel()

	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		t.Parallel()
		schema := services.QuestionImportSchema{
			Text:          "What is 2+2?",
			Type:          "choice",
			Options:       []string{"3", "4", "5", "6"},
			CorrectAnswer: "4",
			Explanation:   "2+2 equals 4",
			Points:        10,
			OrderIndex:    0,
		}

		data, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded services.QuestionImportSchema
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if decoded.Text != "What is 2+2?" {
			t.Errorf("Text = %s, want 'What is 2+2?'", decoded.Text)
		}
		if decoded.Type != "choice" {
			t.Errorf("Type = %s, want 'choice'", decoded.Type)
		}
		if len(decoded.Options) != 4 {
			t.Errorf("Options length = %d, want 4", len(decoded.Options))
		}
		if decoded.Points != 10 {
			t.Errorf("Points = %d, want 10", decoded.Points)
		}
	})

	t.Run("empty options for non-choice types", func(t *testing.T) {
		t.Parallel()
		schema := services.QuestionImportSchema{
			Text:          "Explain photosynthesis",
			Type:          "open",
			Options:       []string{},
			CorrectAnswer: "Photosynthesis is the process...",
			Explanation:   "Plants use sunlight to convert CO2 to glucose",
			Points:        15,
			OrderIndex:    1,
		}

		data, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded services.QuestionImportSchema
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if len(decoded.Options) != 0 {
			t.Errorf("Options length = %d, want 0", len(decoded.Options))
		}
	})

	t.Run("fill type question", func(t *testing.T) {
		t.Parallel()
		schema := services.QuestionImportSchema{
			Text:          "Complete: 2 + 2 = ___",
			Type:          "fill",
			Options:       []string{},
			CorrectAnswer: "4",
			Explanation:   "Basic addition",
			Points:        5,
			OrderIndex:    2,
		}

		data, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded services.QuestionImportSchema
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if decoded.Type != "fill" {
			t.Errorf("Type = %s, want 'fill'", decoded.Type)
		}
		if decoded.Points != 5 {
			t.Errorf("Points = %d, want 5", decoded.Points)
		}
	})
}
