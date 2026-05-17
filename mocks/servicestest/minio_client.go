package servicestest

import (
	"context"
	"io"
	"net/url"
	"reflect"
	"time"

	"github.com/goquizvibe/services"
	"github.com/minio/minio-go/v7"
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

func (m *MockMinioClient) PutObject(ctx context.Context, bucket, object string, r io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutObject", ctx, bucket, object, r, size, opts)
	ret0, _ := ret[0].(minio.UploadInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) PutObject(ctx, bucket, object, r, size, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutObject", reflect.TypeOf((*MockMinioClient)(nil).PutObject), ctx, bucket, object, r, size, opts)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, bucket, object string, opts minio.RemoveObjectOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveObject", ctx, bucket, object, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) RemoveObject(ctx, bucket, object, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveObject", reflect.TypeOf((*MockMinioClient)(nil).RemoveObject), ctx, bucket, object, opts)
}

func (m *MockMinioClient) PresignedGetObject(ctx context.Context, bucket, object string, expires time.Duration, reqParams url.Values) (*url.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PresignedGetObject", ctx, bucket, object, expires, reqParams)
	ret0, _ := ret[0].(*url.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) PresignedGetObject(ctx, bucket, object, expires, reqParams any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PresignedGetObject", reflect.TypeOf((*MockMinioClient)(nil).PresignedGetObject), ctx, bucket, object, expires, reqParams)
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BucketExists", ctx, bucket)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) BucketExists(ctx, bucket any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BucketExists", reflect.TypeOf((*MockMinioClient)(nil).BucketExists), ctx, bucket)
}

func (m *MockMinioClient) MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeBucket", ctx, bucket, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) MakeBucket(ctx, bucket, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeBucket", reflect.TypeOf((*MockMinioClient)(nil).MakeBucket), ctx, bucket, opts)
}

func (m *MockMinioClient) SetBucketPolicy(ctx context.Context, bucket, policy string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBucketPolicy", ctx, bucket, policy)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockMinioClientMockRecorder) SetBucketPolicy(ctx, bucket, policy any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBucketPolicy", reflect.TypeOf((*MockMinioClient)(nil).SetBucketPolicy), ctx, bucket, policy)
}

func (m *MockMinioClient) EndpointURL() *url.URL {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EndpointURL")
	ret0, _ := ret[0].(*url.URL)
	return ret0
}

func (mr *MockMinioClientMockRecorder) EndpointURL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndpointURL", reflect.TypeOf((*MockMinioClient)(nil).EndpointURL))
}

func (m *MockMinioClient) GetObject(ctx context.Context, bucket, object string, opts minio.GetObjectOptions) (*minio.Object, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObject", ctx, bucket, object, opts)
	ret0, _ := ret[0].(*minio.Object)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMinioClientMockRecorder) GetObject(ctx, bucket, object, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObject", reflect.TypeOf((*MockMinioClient)(nil).GetObject), ctx, bucket, object, opts)
}

func (m *MockMinioClient) ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListObjects", ctx, bucket, opts)
	ret0, _ := ret[0].(<-chan minio.ObjectInfo)
	return ret0
}

func (mr *MockMinioClientMockRecorder) ListObjects(ctx, bucket, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListObjects", reflect.TypeOf((*MockMinioClient)(nil).ListObjects), ctx, bucket, opts)
}

var _ services.MinioClient = (*MockMinioClient)(nil)
