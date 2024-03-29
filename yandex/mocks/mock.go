// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/yandex-cloud/terraform-provider-yandex/yandex (interfaces: KafkaTopicModifier)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafka "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

// MockKafkaTopicModifier is a mock of KafkaTopicModifier interface.
type MockKafkaTopicModifier struct {
	ctrl     *gomock.Controller
	recorder *MockKafkaTopicModifierMockRecorder
}

// MockKafkaTopicModifierMockRecorder is the mock recorder for MockKafkaTopicModifier.
type MockKafkaTopicModifierMockRecorder struct {
	mock *MockKafkaTopicModifier
}

// NewMockKafkaTopicModifier creates a new mock instance.
func NewMockKafkaTopicModifier(ctrl *gomock.Controller) *MockKafkaTopicModifier {
	mock := &MockKafkaTopicModifier{ctrl: ctrl}
	mock.recorder = &MockKafkaTopicModifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKafkaTopicModifier) EXPECT() *MockKafkaTopicModifierMockRecorder {
	return m.recorder
}

// CreateKafkaTopic mocks base method.
func (m *MockKafkaTopicModifier) CreateKafkaTopic(arg0 context.Context, arg1 *schema.ResourceData, arg2 *kafka.TopicSpec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateKafkaTopic", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateKafkaTopic indicates an expected call of CreateKafkaTopic.
func (mr *MockKafkaTopicModifierMockRecorder) CreateKafkaTopic(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateKafkaTopic", reflect.TypeOf((*MockKafkaTopicModifier)(nil).CreateKafkaTopic), arg0, arg1, arg2)
}

// DeleteKafkaTopic mocks base method.
func (m *MockKafkaTopicModifier) DeleteKafkaTopic(arg0 context.Context, arg1 *schema.ResourceData, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteKafkaTopic", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteKafkaTopic indicates an expected call of DeleteKafkaTopic.
func (mr *MockKafkaTopicModifierMockRecorder) DeleteKafkaTopic(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteKafkaTopic", reflect.TypeOf((*MockKafkaTopicModifier)(nil).DeleteKafkaTopic), arg0, arg1, arg2)
}

// UpdateKafkaTopic mocks base method.
func (m *MockKafkaTopicModifier) UpdateKafkaTopic(arg0 context.Context, arg1 *schema.ResourceData, arg2 *kafka.TopicSpec, arg3 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateKafkaTopic", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateKafkaTopic indicates an expected call of UpdateKafkaTopic.
func (mr *MockKafkaTopicModifierMockRecorder) UpdateKafkaTopic(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateKafkaTopic", reflect.TypeOf((*MockKafkaTopicModifier)(nil).UpdateKafkaTopic), arg0, arg1, arg2, arg3)
}
