package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/goquizvibe/backend/shared/infrastructure/timeprovider"
	mocks "github.com/goquizvibe/backend/shared/mocks/services"
	servicestest "github.com/goquizvibe/backend/shared/mocks/servicestest"
	"github.com/goquizvibe/backend/shared/models"
	"go.uber.org/mock/gomock"
)

type MockTimeProvider struct {
	now time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.now
}

func NewMockTimeProvider(t time.Time) *MockTimeProvider {
	return &MockTimeProvider{now: t}
}

func NewMockTimeProviderWithNow() *MockTimeProvider {
	return &MockTimeProvider{now: time.Now()}
}

type MockAuthenticator struct {
	ctrl     *gomock.Controller
	recorder *MockAuthenticatorRecorder
}

type MockAuthenticatorRecorder struct {
	mock *MockAuthenticator
}

func NewMockAuthenticator(ctrl *gomock.Controller) *MockAuthenticator {
	m := &MockAuthenticator{ctrl: ctrl}
	m.recorder = &MockAuthenticatorRecorder{mock: m}
	return m
}

func (m *MockAuthenticator) EXPECT() *MockAuthenticatorRecorder {
	return m.recorder
}

func (m *MockAuthenticator) ValidateToken(token string) (*models.AuthClaims, error) {
	ret := m.ctrl.Call(m, "ValidateToken", token)
	ret0, _ := ret[0].(*models.AuthClaims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAuthenticatorRecorder) ValidateToken(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockAuthenticator)(nil).ValidateToken), token)
}

type MockCacheService struct {
	ctrl     *gomock.Controller
	recorder *MockCacheServiceRecorder
}

type MockCacheServiceRecorder struct {
	mock *MockCacheService
}

func NewMockCacheService(ctrl *gomock.Controller) *MockCacheService {
	m := &MockCacheService{ctrl: ctrl}
	m.recorder = &MockCacheServiceRecorder{mock: m}
	return m
}

func (m *MockCacheService) EXPECT() *MockCacheServiceRecorder {
	return m.recorder
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest any) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key, dest)
	return ret[0].(bool)
}

func (mr *MockCacheServiceRecorder) Get(ctx, key, dest any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheService)(nil).Get), ctx, key, dest)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value)
	return ret[0].(error)
}

func (mr *MockCacheServiceRecorder) Set(ctx, key, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheService)(nil).Set), ctx, key, value)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, key)
	return ret[0].(error)
}

func (mr *MockCacheServiceRecorder) Delete(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockCacheService)(nil).Delete), ctx, key)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", ctx, key)
	return ret[0].(bool), ret[1].(error)
}

func (mr *MockCacheServiceRecorder) Exists(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockCacheService)(nil).Exists), ctx, key)
}

func NewTimeProvider(t time.Time) timeprovider.TimeProvider {
	return &MockTimeProvider{now: t}
}

func MustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

func NewRequest(method, urlStr string) *http.Request {
	return httptest.NewRequest(method, urlStr, nil)
}

func NewRequestWithToken(method, urlStr, token string) *http.Request {
	req := httptest.NewRequest(method, urlStr, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
	}
	return req
}

func NewRequestWithCookie(method, urlStr string, cookie *http.Cookie) *http.Request {
	req := httptest.NewRequest(method, urlStr, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	return req
}

type AuthServiceMocks struct {
	Users *mocks.MockUserRepository
}

func NewAuthServiceMocks(t *testing.T) *AuthServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &AuthServiceMocks{
		Users: mocks.NewMockUserRepository(ctrl),
	}
}

func gomockController(t *testing.T) *gomock.Controller {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return ctrl
}

func gomockAny() gomock.Matcher {
	return gomock.Any()
}

type GamificationServiceMocks struct {
	Attempts *mocks.MockAttemptRepository
	Stats    *mocks.MockStatsRepository
}

func NewGamificationServiceMocks(t *testing.T) *GamificationServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &GamificationServiceMocks{
		Attempts: mocks.NewMockAttemptRepository(ctrl),
		Stats:    mocks.NewMockStatsRepository(ctrl),
	}
}

type PermissionsServiceMocks struct {
	Perms         *mocks.MockAssetPermissionRepository
	Groups        *mocks.MockUserGroupRepository
	StudentAccess *mocks.MockStudentAccessRepository
}

func NewPermissionsServiceMocks(t *testing.T) *PermissionsServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &PermissionsServiceMocks{
		Perms:         mocks.NewMockAssetPermissionRepository(ctrl),
		Groups:        mocks.NewMockUserGroupRepository(ctrl),
		StudentAccess: mocks.NewMockStudentAccessRepository(ctrl),
	}
}

type QuizServiceMocks struct {
	Quizzes   *mocks.MockQuizRepository
	Questions *mocks.MockQuestionRepository
	Images    *mocks.MockImageRepository
	Attempts  *mocks.MockAttemptRepository
	Groups    *mocks.MockUserGroupRepository
}

func NewQuizServiceMocks(t *testing.T) *QuizServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &QuizServiceMocks{
		Quizzes:   mocks.NewMockQuizRepository(ctrl),
		Questions: mocks.NewMockQuestionRepository(ctrl),
		Images:    mocks.NewMockImageRepository(ctrl),
		Attempts:  mocks.NewMockAttemptRepository(ctrl),
		Groups:    mocks.NewMockUserGroupRepository(ctrl),
	}
}

type QuizSessionServiceMocks struct {
	Attempts  *mocks.MockAttemptRepository
	Quizzes   *mocks.MockQuizRepository
	Questions *mocks.MockQuestionRepository
	Images    *mocks.MockImageRepository
	Users     *mocks.MockUserRepository
	Stats     *mocks.MockStatsRepository
}

func NewQuizSessionServiceMocks(t *testing.T) *QuizSessionServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &QuizSessionServiceMocks{
		Attempts:  mocks.NewMockAttemptRepository(ctrl),
		Quizzes:   mocks.NewMockQuizRepository(ctrl),
		Questions: mocks.NewMockQuestionRepository(ctrl),
		Images:    mocks.NewMockImageRepository(ctrl),
		Users:     mocks.NewMockUserRepository(ctrl),
		Stats:     mocks.NewMockStatsRepository(ctrl),
	}
}

type DashboardServiceMocks struct {
	Users     *mocks.MockUserRepository
	Quizzes   *mocks.MockQuizRepository
	Questions *mocks.MockQuestionRepository
	Images    *mocks.MockImageRepository
	Attempts  *mocks.MockAttemptRepository
	Stats     *mocks.MockStatsRepository
	Groups    *mocks.MockUserGroupRepository
}

func NewDashboardServiceMocks(t *testing.T) *DashboardServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &DashboardServiceMocks{
		Users:     mocks.NewMockUserRepository(ctrl),
		Quizzes:   mocks.NewMockQuizRepository(ctrl),
		Questions: mocks.NewMockQuestionRepository(ctrl),
		Images:    mocks.NewMockImageRepository(ctrl),
		Attempts:  mocks.NewMockAttemptRepository(ctrl),
		Stats:     mocks.NewMockStatsRepository(ctrl),
		Groups:    mocks.NewMockUserGroupRepository(ctrl),
	}
}

type AdminServiceMocks struct {
	Users       *mocks.MockUserRepository
	Quizzes     *mocks.MockQuizRepository
	Questions   *mocks.MockQuestionRepository
	Images      *mocks.MockImageRepository
	Attempts    *mocks.MockAttemptRepository
	Stats       *mocks.MockStatsRepository
	Materials   *mocks.MockLearningMaterialRepository
	Permissions *mocks.MockAssetPermissionRepository
}

func NewAdminServiceMocks(t *testing.T) *AdminServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &AdminServiceMocks{
		Users:       mocks.NewMockUserRepository(ctrl),
		Quizzes:     mocks.NewMockQuizRepository(ctrl),
		Questions:   mocks.NewMockQuestionRepository(ctrl),
		Images:      mocks.NewMockImageRepository(ctrl),
		Attempts:    mocks.NewMockAttemptRepository(ctrl),
		Stats:       mocks.NewMockStatsRepository(ctrl),
		Materials:   mocks.NewMockLearningMaterialRepository(ctrl),
		Permissions: mocks.NewMockAssetPermissionRepository(ctrl),
	}
}

type StorageServiceMocks struct {
	Minio *servicestest.MockMinioClient
}

func NewStorageServiceMocks(t *testing.T) *StorageServiceMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return &StorageServiceMocks{
		Minio: servicestest.NewMockMinioClient(ctrl),
	}
}
