package xerrors

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Issue"

	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

const (
	IssueCodeTransactionLocksInvalidated       = 2001
	IssueCodeDatashardProgramSizeLimitExceeded = 200509
)

type issues []*Ydb_Issue.IssueMessage

func (ii issues) String() string {
	if len(ii) == 0 {
		return ""
	}
	b := xstring.Buffer()
	defer b.Free()
	b.WriteByte('[')
	for i, m := range ii {
		if i != 0 {
			b.WriteByte(',')
		}
		b.WriteByte('{')
		if p := m.GetPosition(); p != nil {
			if file := p.GetFile(); file != "" {
				b.WriteString(file)
				b.WriteByte(':')
			}
			b.WriteString(strconv.Itoa(int(p.GetRow())))
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(int(p.GetColumn())))
			b.WriteString(" => ")
		}
		if code := m.GetIssueCode(); code != 0 {
			b.WriteByte('#')
			b.WriteString(strconv.Itoa(int(code)))
			b.WriteByte(' ')
		}
		b.WriteByte('\'')
		b.WriteString(strings.TrimSuffix(m.GetMessage(), "."))
		b.WriteByte('\'')
		if len(m.GetIssues()) > 0 {
			b.WriteByte(' ')
			b.WriteString(issues(m.GetIssues()).String())
		}
		b.WriteByte('}')
	}
	b.WriteByte(']')

	return b.String()
}

// NewWithIssues returns error which contains child issues
func NewWithIssues(text string, issues ...error) error {
	err := &withIssuesError{
		reason: text,
	}
	for i := range issues {
		if issues[i] != nil {
			err.issues = append(err.issues, issues[i])
		}
	}

	return err
}

type withIssuesError struct {
	reason string
	issues []error
}

func (e *withIssuesError) isYdbError() {}

func (e *withIssuesError) Error() string {
	b := xstring.Buffer()
	defer b.Free()
	if len(e.reason) > 0 {
		b.WriteString(e.reason)
		b.WriteString(", issues: [")
	} else {
		b.WriteString("multiple errors: [")
	}
	for i, issue := range e.issues {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(issue.Error())
	}
	b.WriteString("]")

	return b.String()
}

func (e *withIssuesError) As(target interface{}) bool {
	for _, err := range e.issues {
		if As(err, target) {
			return true
		}
	}

	return false
}

func (e *withIssuesError) Is(target error) bool {
	for _, err := range e.issues {
		if Is(err, target) {
			return true
		}
	}

	return false
}

func (e *withIssuesError) Unwrap() []error {
	return e.issues
}

// Issue struct
type Issue struct {
	Message  string
	Code     uint32
	Severity uint32
}

type IssueIterator []*Ydb_Issue.IssueMessage

func (it IssueIterator) Len() int {
	return len(it)
}

func (it IssueIterator) Get(i int) (issue Issue, nested IssueIterator) {
	x := it[i]
	if xs := x.GetIssues(); len(xs) > 0 {
		nested = IssueIterator(xs)
	}

	return Issue{
		Message:  x.GetMessage(),
		Code:     x.GetIssueCode(),
		Severity: x.GetSeverity(),
	}, nested
}

func IterateByIssues(err error, it func(message string, code Ydb.StatusIds_StatusCode, severity uint32)) {
	_ = iterateByIssues(err, func(message string, code Ydb.StatusIds_StatusCode, severity uint32) (stop bool) {
		it(message, code, severity)

		return false
	})
}

func iterateByIssues(err error,
	it func(message string, code Ydb.StatusIds_StatusCode, severity uint32) (stop bool),
) (stopped bool) {
	var o *operationError
	if !errors.As(err, &o) {
		return
	}

	return iterate(o.Issues(), it)
}

func iterate(
	issues []*Ydb_Issue.IssueMessage,
	it func(message string, code Ydb.StatusIds_StatusCode, severity uint32) (stop bool),
) (stopped bool) {
	for _, issue := range issues {
		if it(issue.GetMessage(), Ydb.StatusIds_StatusCode(issue.GetIssueCode()), issue.GetSeverity()) {
			return true
		}

		if iterate(issue.GetIssues(), it) {
			return true
		}
	}

	return false
}
