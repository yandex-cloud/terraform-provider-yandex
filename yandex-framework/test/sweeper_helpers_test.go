package test

import "testing"

func Test_testResourseName(t *testing.T) {
	t.Parallel()

	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name:    "case: base case",
			args:    args{length: 63},
			wantLen: 63,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(testResourseName(tt.args.length)); got != tt.wantLen {
				t.Errorf("testResourseName() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
