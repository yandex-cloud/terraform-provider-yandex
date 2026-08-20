package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

func TestSuccessfulCreateRetry(t *testing.T) {
	alreadyOwned := awserr.New("BucketAlreadyOwnedByYou", "already created", nil)
	if !isSuccessfulCreateRetry(alreadyOwned, 1) {
		t.Fatal("expected an already-owned response after a retry to be accepted")
	}
	if isSuccessfulCreateRetry(alreadyOwned, 0) {
		t.Fatal("must not accept an already-owned response on the first attempt")
	}
	if isSuccessfulCreateRetry(awserr.New("AccessDenied", "denied", nil), 1) {
		t.Fatal("must not accept a different error after a retry")
	}
}
