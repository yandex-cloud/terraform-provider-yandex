package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getSdkPath(t *testing.T) {
	t.Parallel()

	type args struct {
		service  string
		resource string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case: positive case",
			args: args{
				service:  "datasphere",
				resource: "community",
			},
			want: "SDK.Datasphere().Community()",
		},
	}
	for _, tt := range tests {

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Arrange, Act
			path := getSdkPath(tt.args.service, tt.args.resource)

			// Assert
			assert.Equal(t, tt.want, path)
		})
	}
}
