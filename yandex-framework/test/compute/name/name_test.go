package name

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestIsTestResource(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case: should return true if args.name has test prefix",
			args: args{name: fmt.Sprintf("%s-%s", TestPrefix(), acctest.RandString(10))},
			want: true,
		},
		{
			name: "case: should return false if args.name doesn't have test prefix",
			args: args{name: fmt.Sprintf("%s-%s", "blablabla", acctest.RandString(10))},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsTestResource(tt.args.name); got != tt.want {
				t.Errorf("IsTestResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateNameForResource(t *testing.T) {
	t.Parallel()

	type args struct {
		suffixLen int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name:    "case: should return string with the length == prefix.len + args.suffixLen + 1",
			args:    args{suffixLen: 10},
			wantLen: len(testPrefix) + 10 + 1,
		},
	}
	for _, tt := range tests {

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := GenerateNameForResource(tt.args.suffixLen); len(got) != tt.wantLen {
				t.Errorf("GenerateNameForResource() len = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
