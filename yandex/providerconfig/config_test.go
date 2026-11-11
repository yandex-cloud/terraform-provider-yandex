package providerconfig

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_iamKeyV2FromJSONContent(t *testing.T) {
	content, err := ioutil.ReadFile("../test-fixtures/fake_service_account_key.json")
	require.NoError(t, err, "fail on file read %s", "../test-fixtures/fake_service_account_key.json")
	_, err = iamKeyV2FromJSONContent(string(content))
	require.NoError(t, err)
}
