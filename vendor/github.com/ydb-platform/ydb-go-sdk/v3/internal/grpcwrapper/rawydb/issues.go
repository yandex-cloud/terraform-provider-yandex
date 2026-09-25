package rawydb

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Issue"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type Issues []Issue

func (issuesPointer *Issues) FromProto(p []*Ydb_Issue.IssueMessage) error {
	*issuesPointer = make(Issues, len(p))
	issues := *issuesPointer
	for i := range issues {
		if err := issues[i].FromProto(p[i]); err != nil {
			return err
		}
	}

	return nil
}

func (issuesPointer *Issues) String() string {
	issues := *issuesPointer
	issuesStrings := make([]string, len(issues))
	for i := range issues {
		issuesStrings[i] = issues[i].String()
	}

	return strings.Join(issuesStrings, ", ")
}

type Issue struct {
	Code    uint32
	Message string
	Issues  Issues
}

func (issue *Issue) FromProto(p *Ydb_Issue.IssueMessage) error {
	if p == nil {
		return xerrors.WithStackTrace(errors.New("receive nil issue message pointer from protobuf"))
	}
	issue.Code = p.GetIssueCode()
	issue.Message = p.GetMessage()

	return issue.Issues.FromProto(p.GetIssues())
}

func (issue *Issue) String() string {
	var innerIssues string
	if len(issue.Issues) > 0 {
		innerIssues = " (" + issue.Issues.String() + ")"
	}

	return fmt.Sprintf("message: %v, code: %v%v", issue.Message, issue.Code, innerIssues)
}

// Equals compares this Issue with another Issue for equality
func (issue *Issue) Equals(other *Issue) bool {
	if issue == nil && other == nil {
		return true
	}
	if issue == nil || other == nil {
		return false
	}

	if issue.Code != other.Code {
		return false
	}
	if issue.Message != other.Message {
		return false
	}

	return issue.Issues.Equals(other.Issues)
}

// Equals compares this Issues slice with another Issues slice for equality
func (issuesPointer Issues) Equals(other Issues) bool {
	if len(issuesPointer) != len(other) {
		return false
	}

	for i := range issuesPointer {
		if !issuesPointer[i].Equals(&other[i]) {
			return false
		}
	}

	return true
}
