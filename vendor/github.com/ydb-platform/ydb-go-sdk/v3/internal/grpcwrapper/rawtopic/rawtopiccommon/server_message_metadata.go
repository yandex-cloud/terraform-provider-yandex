package rawtopiccommon

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Issue"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

type StatusAndIssues interface {
	GetStatus() Ydb.StatusIds_StatusCode
	GetIssues() []*Ydb_Issue.IssueMessage
}

type ServerMessageMetadata struct {
	Status rawydb.StatusCode
	Issues rawydb.Issues
}

func (m *ServerMessageMetadata) MetaFromStatusAndIssues(p StatusAndIssues) error {
	if err := m.Status.FromProto(p.GetStatus()); err != nil {
		return err
	}

	return m.Issues.FromProto(p.GetIssues())
}

func (m *ServerMessageMetadata) StatusData() ServerMessageMetadata {
	return *m
}

func (m *ServerMessageMetadata) SetStatus(status rawydb.StatusCode) {
	m.Status = status
}

// Equals compares this ServerMessageMetadata with another ServerMessageMetadata for equality
func (m *ServerMessageMetadata) Equals(other *ServerMessageMetadata) bool {
	if m == nil && other == nil {
		return true
	}
	if m == nil || other == nil {
		return false
	}

	// Compare status using StatusCode's Equals method
	if !m.Status.Equals(other.Status) {
		return false
	}

	// Compare issues using Issues' Equals method
	if !m.Issues.Equals(other.Issues) {
		return false
	}

	return true
}
