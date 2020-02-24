// Code generated by MockGen. DO NOT EDIT.
// Source: micro/store/store.go

// Package mock_store is a generated GoMock package.
package mock_store

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	store "go.zoe.im/x/micro/store"
	reflect "reflect"
	time "time"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// List mocks base method
func (m *MockStore) List(ctx context.Context, ops ...store.ListOption) ([]store.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range ops {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].([]store.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockStoreMockRecorder) List(ctx interface{}, ops ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, ops...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockStore)(nil).List), varargs...)
}

// Get mocks base method
func (m *MockStore) Get(ctx context.Context, id string, ops ...store.GetOption) (store.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, id}
	for _, a := range ops {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(store.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockStoreMockRecorder) Get(ctx, id interface{}, ops ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, id}, ops...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStore)(nil).Get), varargs...)
}

// Create mocks base method
func (m *MockStore) Create(ctx context.Context, r store.Record, ops ...store.CreateOption) (store.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, r}
	for _, a := range ops {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(store.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockStoreMockRecorder) Create(ctx, r interface{}, ops ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, r}, ops...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStore)(nil).Create), varargs...)
}

// Update mocks base method
func (m *MockStore) Update(ctx context.Context, r store.Record, ops ...store.UpdateOption) (store.Record, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, r}
	for _, a := range ops {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(store.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update
func (mr *MockStoreMockRecorder) Update(ctx, r interface{}, ops ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, r}, ops...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockStore)(nil).Update), varargs...)
}

// Delete mocks base method
func (m *MockStore) Delete(ctx context.Context, id string, ops ...store.DeleteOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, id}
	for _, a := range ops {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockStoreMockRecorder) Delete(ctx, id interface{}, ops ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, id}, ops...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStore)(nil).Delete), varargs...)
}

// MockRecord is a mock of Record interface
type MockRecord struct {
	ctrl     *gomock.Controller
	recorder *MockRecordMockRecorder
}

// MockRecordMockRecorder is the mock recorder for MockRecord
type MockRecordMockRecorder struct {
	mock *MockRecord
}

// NewMockRecord creates a new mock instance
func NewMockRecord(ctrl *gomock.Controller) *MockRecord {
	mock := &MockRecord{ctrl: ctrl}
	mock.recorder = &MockRecordMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRecord) EXPECT() *MockRecordMockRecorder {
	return m.recorder
}

// GetID mocks base method
func (m *MockRecord) GetID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetID indicates an expected call of GetID
func (mr *MockRecordMockRecorder) GetID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockRecord)(nil).GetID))
}

// GetData mocks base method
func (m *MockRecord) GetData() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetData")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetData indicates an expected call of GetData
func (mr *MockRecordMockRecorder) GetData() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetData", reflect.TypeOf((*MockRecord)(nil).GetData))
}

// GetValue mocks base method
func (m *MockRecord) GetValue() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValue")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// GetValue indicates an expected call of GetValue
func (mr *MockRecordMockRecorder) GetValue() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValue", reflect.TypeOf((*MockRecord)(nil).GetValue))
}

// GetExpiry mocks base method
func (m *MockRecord) GetExpiry() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExpiry")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// GetExpiry indicates an expected call of GetExpiry
func (mr *MockRecordMockRecorder) GetExpiry() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExpiry", reflect.TypeOf((*MockRecord)(nil).GetExpiry))
}

// GetCreateAt mocks base method
func (m *MockRecord) GetCreateAt() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreateAt")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetCreateAt indicates an expected call of GetCreateAt
func (mr *MockRecordMockRecorder) GetCreateAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreateAt", reflect.TypeOf((*MockRecord)(nil).GetCreateAt))
}

// GetUpdateAt mocks base method
func (m *MockRecord) GetUpdateAt() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdateAt")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetUpdateAt indicates an expected call of GetUpdateAt
func (mr *MockRecordMockRecorder) GetUpdateAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdateAt", reflect.TypeOf((*MockRecord)(nil).GetUpdateAt))
}

// GetDeleteAt mocks base method
func (m *MockRecord) GetDeleteAt() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeleteAt")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetDeleteAt indicates an expected call of GetDeleteAt
func (mr *MockRecordMockRecorder) GetDeleteAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeleteAt", reflect.TypeOf((*MockRecord)(nil).GetDeleteAt))
}
