package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/goquizvibe/db"
)

type UserRepository interface {
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	GetStudentCount(ctx context.Context) (int64, error)
}

type QuizRepository interface {
	CreateQuiz(ctx context.Context, params db.CreateQuizParams) (db.Quiz, error)
	GetQuizByID(ctx context.Context, id uuid.UUID) (db.Quiz, error)
	GetQuizzesForUser(ctx context.Context, userID uuid.UUID) ([]db.Quiz, error)
	GetNonArchivedQuizzes(ctx context.Context) ([]db.Quiz, error)
	UpdateQuiz(ctx context.Context, params db.UpdateQuizParams) (db.Quiz, error)
	UpdateQuizStatus(ctx context.Context, params db.UpdateQuizStatusParams) error
	DeleteQuiz(ctx context.Context, id uuid.UUID) error
}

type QuestionRepository interface {
	CreateQuestion(ctx context.Context, params db.CreateQuestionParams) (db.Question, error)
	GetQuestionByID(ctx context.Context, id uuid.UUID) (db.Question, error)
	GetQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) ([]db.Question, error)
	UpdateQuestion(ctx context.Context, params db.UpdateQuestionParams) (db.Question, error)
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	GetMaxOrderIndex(ctx context.Context, quizID uuid.UUID) (interface{}, error)
}

type AttemptRepository interface {
	CreateAttempt(ctx context.Context, params db.CreateAttemptParams) (db.QuizAttempt, error)
	UpdateAttempt(ctx context.Context, params db.UpdateAttemptParams) (db.QuizAttempt, error)
	GetAttemptByID(ctx context.Context, id uuid.UUID) (db.QuizAttempt, error)
	GetAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]db.QuizAttempt, error)
	GetIncompleteAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]db.QuizAttempt, error)
	GetAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]db.UserAnswer, error)
	UpsertUserAnswer(ctx context.Context, params db.UpsertUserAnswerParams) (db.UserAnswer, error)
	GetQuizErrors(ctx context.Context, userID uuid.UUID) ([]db.QuizAttempt, error)
	GetRecentAttempts(ctx context.Context, limit int32) ([]db.GetRecentAttemptsRow, error)
	GetStaleAttempts(ctx context.Context) ([]db.GetStaleAttemptsRow, error)
}

type ImageRepository interface {
	CreateQuestionImage(ctx context.Context, params db.CreateQuestionImageParams) (db.QuestionImage, error)
	GetQuestionImageByID(ctx context.Context, id uuid.UUID) (db.QuestionImage, error)
	GetImagesByQuestionID(ctx context.Context, questionID uuid.UUID) ([]db.QuestionImage, error)
	GetImageCountByQuestionID(ctx context.Context, questionID uuid.UUID) (int64, error)
	DeleteQuestionImage(ctx context.Context, id uuid.UUID) error
}

type StatsRepository interface {
	GetUserStats(ctx context.Context, userID uuid.UUID) (db.GetUserStatsRow, error)
	GetAdminStatsData(ctx context.Context) (db.GetAdminStatsDataRow, error)
	GetQuizStats(ctx context.Context) ([]db.GetQuizStatsRow, error)
	GetGradeDistribution(ctx context.Context) ([]byte, error)
	GetSubjectDistribution(ctx context.Context) ([]byte, error)
	GetLastActiveDate(ctx context.Context, userID uuid.UUID) (any, error)
}

type LearningMaterialRepository interface {
	CreateLearningMaterial(ctx context.Context, params db.CreateLearningMaterialParams) (db.LearningMaterial, error)
	GetLearningMaterialByID(ctx context.Context, id uuid.UUID) (db.LearningMaterial, error)
	GetLearningMaterials(ctx context.Context) ([]db.LearningMaterial, error)
	GetLearningMaterialsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.LearningMaterial, error)
	UpdateLearningMaterial(ctx context.Context, params db.UpdateLearningMaterialParams) (db.LearningMaterial, error)
	DeleteLearningMaterial(ctx context.Context, id uuid.UUID) error
}
