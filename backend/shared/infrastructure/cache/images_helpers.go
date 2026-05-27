package services

import (
	"context"

	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/models"
	r "github.com/goquizvibe/backend/shared/repositories"
)

func AttachImagesToQuestions(ctx context.Context, questions []db.Question, imagesRepo r.ImageRepository) []models.QuestionWithImages {
	result := make([]models.QuestionWithImages, len(questions))
	for i, q := range questions {
		images, _ := imagesRepo.GetImagesByQuestionID(ctx, q.ID)
		result[i] = models.QuestionWithImages{
			Question: q,
			Images:   images,
		}
	}
	return result
}