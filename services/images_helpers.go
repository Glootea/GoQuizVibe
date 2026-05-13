package services

import (
	"context"

	"github.com/goquizvibe/db"
	"github.com/goquizvibe/models"
	r "github.com/goquizvibe/repositories"
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