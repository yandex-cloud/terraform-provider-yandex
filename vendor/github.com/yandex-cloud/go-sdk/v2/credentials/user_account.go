package credentials

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/yandex-cloud/go-sdk/v2/pkg/iamkey"
)

// UserAccountKey returns credentials for the given IAM Key. The key is used to sign JWT tokens.
// JWT tokens are exchanged for IAM Tokens used to authorize API calls.
//
// WARN: user account keys are not supported, and won't be supported for most users.
func UserAccountKey(key *iamkey.Key) (ExchangeableCredentials, error) {
	userAccountID := key.GetUserAccountId()
	if userAccountID == "" {
		return nil, fmt.Errorf("key should de issued for user account, but subject is %#v", key.Subject)
	}

	// User account key usage is same as service account key.
	key = proto.Clone(key).(*iamkey.Key)
	key.Subject = &iamkey.Key_ServiceAccountId{ServiceAccountId: userAccountID}

	return ServiceAccountKey(key)
}

func UserAccountKeyFile(keyFilePath string) (Credentials, error) {
	key, err := iamkey.ReadFromJSONFile(keyFilePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to load service account key from %s", keyFilePath)
	}

	return UserAccountKey(key)
}
