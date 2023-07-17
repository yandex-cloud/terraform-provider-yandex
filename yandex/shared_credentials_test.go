package yandex

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestRetrieve(t *testing.T) {
	profile := "prod-profile"
	cases := []struct {
		name                string
		fileContent         string
		expectedCredentials *SharedCredentials
		expectedError       string
	}{
		{
			name:          "empty file content",
			fileContent:   "",
			expectedError: "not found shared credentials for `prod-profile` profile",
		},
		{
			name:          "unsupported shared credentials file format",
			fileContent:   "data",
			expectedError: "unsupported shared credentials file format, line `data` does not match",
		},
		{
			name:                "valid shared credentials file with single profile",
			fileContent:         "[prod-profile]\nstorage_access_key=access-key\nstorage_secret_key=secret-key",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "access-key", StorageSecretKey: "secret-key"},
		},
		{
			name: "valid shared credentials file with multiple profiles",
			fileContent: "[dev-profile]\nstorage_access_key=dev-access-key\nstorage_secret_key=dev-secret-key\n\n" +
				"[prod-profile]\nstorage_access_key=prod-access-key\nstorage_secret_key=prod-secret-key",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "prod-access-key", StorageSecretKey: "prod-secret-key"},
		},
		{
			name:          "credentials without profile",
			fileContent:   "storage_access_key=dev-access-key\nstorage_secret_key=dev-secret-key",
			expectedError: "key `storage_access_key` does not have a profile",
		},
		{
			name:          "duplicate credentials",
			fileContent:   "[prod-profile]\nstorage_access_key=dev-access-key\nstorage_access_key=dev-access-key",
			expectedError: "key `storage_access_key` has multiple values",
		},
		{
			name:                "multiple equality symbols",
			fileContent:         "[prod-profile]\nstorage_access_key=access-key===\nstorage_secret_key=secret-key===",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "access-key===", StorageSecretKey: "secret-key==="},
		},
		{
			name: "extra credentials",
			fileContent: "[prod-profile]\nstorage_access_key=access-key\nstorage_secret_key=secret-key\n" +
				"extra_key=extra-key",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "access-key", StorageSecretKey: "secret-key"},
		},
		{
			name:                "partial credentials",
			fileContent:         "[prod-profile]\nstorage_access_key=access-key",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "access-key", StorageSecretKey: ""},
		},
		{
			name:                "no credentials",
			fileContent:         "[prod-profile]",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "", StorageSecretKey: ""},
		},
		{
			name:          "no given profile",
			fileContent:   "[testing-profile]\nstorage_access_key=access-key\nstorage_secret_key=secret-key",
			expectedError: "not found shared credentials for `prod-profile` profile",
		},
		{
			name:                "trim key/value empty spaces",
			fileContent:         "[prod-profile]\n storage_access_key  = access-key  \n storage_secret_key  = secret-key  ",
			expectedCredentials: &SharedCredentials{StorageAccessKey: "access-key", StorageSecretKey: "secret-key"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filename, err := writeFile(tc.fileContent)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(filename)
			sharedCredentialsFileProvider := SharedCredentialsProvider{filename, profile}

			result, err := sharedCredentialsFileProvider.Retrieve()
			if err != nil {
				if tc.expectedError == "" {
					t.Errorf("Unexpected error: %#v", err.Error())
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected `%v` error, but got `%v`", tc.expectedError, err.Error())
				}
			} else {
				assert.Equal(t, tc.expectedCredentials, result)
			}
		})
	}
}

func writeFile(data string) (string, error) {
	tmpFile, err := os.CreateTemp("", "shared-credentials-file-test")

	if err != nil {
		return "", err
	}

	err = os.WriteFile(tmpFile.Name(), []byte(data), 0777)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), err
}
