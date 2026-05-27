package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/goquizvibe/backend/shared/db"
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
	GetRecentLearningMaterials(ctx context.Context, limit int32) ([]db.LearningMaterial, error)
}

type UserGroupRepository interface {
	CreateUserGroup(ctx context.Context, params db.CreateUserGroupParams) (db.UserGroup, error)
	GetUserGroupByID(ctx context.Context, id uuid.UUID) (db.UserGroup, error)
	GetUserGroupsByAdmin(ctx context.Context, userID uuid.UUID) ([]db.UserGroup, error)
	UpdateUserGroup(ctx context.Context, params db.UpdateUserGroupParams) (db.UserGroup, error)
	DeleteUserGroup(ctx context.Context, params db.DeleteUserGroupParams) error
	AddUserToGroup(ctx context.Context, params db.AddUserToGroupParams) (db.GroupMember, error)
	RemoveUserFromGroup(ctx context.Context, params db.RemoveUserFromGroupParams) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]db.GetGroupMembersRow, error)
	GetUserRoleInGroup(ctx context.Context, params db.GetUserRoleInGroupParams) (interface{}, error)
	GetGroupMemberCount(ctx context.Context, groupID uuid.UUID) (int64, error)
	IsUserMemberOfGroup(ctx context.Context, params db.IsUserMemberOfGroupParams) (bool, error)
}

type AssetPermissionRepository interface {
	SetOwnerPermission(ctx context.Context, params db.SetOwnerPermissionParams) (db.AssetPermission, error)
	GrantPermission(ctx context.Context, params db.GrantPermissionParams) (db.AssetPermission, error)
	RevokePermission(ctx context.Context, params db.RevokePermissionParams) error
	GetAssetPermissions(ctx context.Context, params db.GetAssetPermissionsParams) ([]db.GetAssetPermissionsRow, error)
	GetUserAssetPermissions(ctx context.Context, recipientID uuid.UUID) ([]db.AssetPermission, error)
	GetGroupAssetPermissions(ctx context.Context, groupIDs []uuid.UUID) ([]db.AssetPermission, error)
	GetAccessibleAssetIDs(ctx context.Context, params db.GetAccessibleAssetIDsParams) ([]uuid.UUID, error)
	DeleteAssetPermissionsByAsset(ctx context.Context, params db.DeleteAssetPermissionsByAssetParams) error
	HasPermission(ctx context.Context, params db.HasPermissionParams) (bool, error)
	HasPermissionLevel(ctx context.Context, params db.HasPermissionLevelParams) (bool, error)
}
