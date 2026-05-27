package servicestest

import (
	"context"
	"reflect"
	"time"

	"github.com/goquizvibe/pkg/storage"
	"go.uber.org/mock/gomock"
)

type MockMinioClient struct {
	ctrl     *gomock.Controller
	recorder *MockMinioClientMockRecorder
	isgomock struct{}
}

type MockMinioClientMockRecorder struct {
	mock *MockMinioClient
}

func NewMockMinioClient(ctrl *gomock.Controller) *MockMinioClient {
	mock := &MockMinioClient{ctrl: ctrl}
	mock.recorder = &MockMinioClientMockRecorder{mock: mock}
	return mock
}

func (m *MockMinioClient) EXPECT() *MockMinioClientMockRecorder {
	return m.recorder
}

func (m *MockMinioClient) PutObject(ctx context.Context, path string, data []byte, contentType string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutObject", ctx, path, data, contentType)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) PutObject(ctx, path, data, contentType any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutObject", reflect.TypeOf((*MockMinioClient)(nil).PutObject), ctx, path, data, contentType)
}

func (m *MockMinioClient) GetObject(ctx context.Context, path string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObject", ctx, path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) GetObject(ctx, path any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObject", reflect.TypeOf((*MockMinioClient)(nil).GetObject), ctx, path)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveObject", ctx, path)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) RemoveObject(ctx, path any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveObject", reflect.TypeOf((*MockMinioClient)(nil).RemoveObject), ctx, path)
}

func (m *MockMinioClient) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjects", ctx, prefix)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) ListObjects(ctx, prefix any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjects", reflect.TypeOf((*MockMinioClient)(nil).ListObjects), ctx, prefix)
}

func (m *MockMinioClient) EnsureBucket(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureBucket", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) EnsureBucket(ctx any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureBucket", reflect.TypeOf((*MockMinioClient)(nil).EnsureBucket), ctx)
}

func (m *MockMinioClient) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPresignedURL", ctx, path, expiry)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) GetPresignedURL(ctx, path, expiry any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPresignedURL", reflect.TypeOf((*MockMinioClient)(nil).GetPresignedURL), ctx, path, expiry)
}

func (m *MockMinioClient) Bucket() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bucket")
	ret0, _ := ret[0].(string)
	return ret0
}

func (mr *MockMinioClientMockRecorder) Bucket() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bucket", reflect.TypeOf((*MockMinioClient)(nil).Bucket))
}

func (m *MockMinioClient) Endpoint() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Endpoint")
	ret0, _ := ret[0].(string)
	return ret0
}

func (mr *MockMinioClientMockRecorder) Endpoint() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Endpoint", reflect.TypeOf((*MockMinioClient)(nil).Endpoint))
}

func (m *MockMinioClient) SetBucketPolicy(ctx context.Context, bucket, policy string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBucketPolicy", ctx, bucket, policy)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) SetBucketPolicy(ctx, bucket, policy any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBucketPolicy", reflect.TypeOf((*MockMinioClient)(nil).SetBucketPolicy), ctx, bucket, policy)
}

var _ storage.Storage = (*MockMinioClient)(nil)