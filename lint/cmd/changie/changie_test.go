package main

import (
	"testing"
)

func Test_validateChangieBody(t *testing.T) {
	t.Parallel()

	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case: if the changie description is empty, must return false",
			args: args{input: ""},
			want: false,
		},
		{
			name: "case: if the changie description has wrong format, must return false",
			args: args{input: "Add support for attach detach instance network interfaces"},
			want: false,
		},
		{
			name: "case: if the changie description has wrong format, must return false",
			args: args{input: "ALB:Add support for attach detach instance network interfaces"},
			want: false,
		},
		{
			name: "case: if the changie description has wrong format, must return false",
			args: args{input: "ALB"},
			want: false,
		},
		{
			name: "case: if the changie description has wrong format, must return false",
			args: args{input: "ALB:"},
			want: false,
		},
		{
			name: "case: if the changie description has wrong format, must return false",
			args: args{input: "ALB: "},
			want: false,
		},
		{
			name: "case: if the changie description has valid format, must return false",
			args: args{input: "postgres_cluster:Add support for `attach` detach instance network interfaces"},
			want: false,
		},
		{
			name: "case: if the changie description has valid format, must return false",
			args: args{input: " postgres_cluster: Add support for `attach` detach instance network interfaces"},
			want: false,
		},
		{
			name: "case: if the changie description has valid format, must return false",
			args: args{input: "postgres_cluster : Add support for `attach` detach instance network interfaces"},
			want: false,
		},
		{
			name: "case: if the changie description has valid format, must return true",
			args: args{input: "ALB: Add support for attach detach instance network interfaces"},
			want: true,
		},
		{
			name: "case: if the changie description has valid format, must return true",
			args: args{input: "alb: Add support for attach detach instance network interfaces"},
			want: true,
		},
		{
			name: "case: if the changie description has valid format, must return true",
			args: args{input: "postgres_cluster: Add support for attach detach instance network interfaces"},
			want: true,
		},
		{
			name: "case: if the changie description has valid format, must return true",
			args: args{input: "postgres_cluster: Add support for `attach` detach instance network interfaces"},
			want: true,
		},
	}
	for _, tt := range tests {

		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := validateChangieBody(tt.args.input); got != tt.want {
				t.Errorf("validateChangieBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleYamlFile(t *testing.T) {
	t.Parallel()

	type args struct {
		content []byte
		unrel   *unreleased
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case: must not return error if unreleased changie matches the format",
			args: args{
				content: []byte(`
kind: ENHANCEMENTS
body: "clickhouse: save implicit zookeeper subcluster on hosts changes"
time: 2024-06-21T13:19:18.006693+02:00`,
				),
				unrel: &unreleased{},
			},
			wantErr: false,
		},
		{
			name: "case: must not return error if unreleased changie matches the format",
			args: args{
				content: []byte(`
kind: ENHANCEMENTS
body: "clickhouse_cluster: save implicit zookeeper subcluster on hosts changes."
time: 2024-06-21T13:19:18.006693+02:00`,
				),
				unrel: &unreleased{},
			},
			wantErr: false,
		},
		{
			name: "case: must return error if unreleased changie doesn't matches the format",
			args: args{
				content: []byte(`
kind: ENHANCEMENTS
body: "save implicit zookeeper subcluster on hosts changes."
time: 2024-06-21T13:19:18.006693+02:00`,
				),
				unrel: &unreleased{},
			},
			wantErr: true,
		},
		{
			name: "case: must return error if unreleased changie doesn't matches the format",
			args: args{
				content: []byte(`
kind: ENHANCEMENTS
body: ""
time: 2024-06-21T13:19:18.006693+02:00`,
				),
				unrel: &unreleased{},
			},
			wantErr: true,
		},
		{
			name: "case: must return error if unreleased changie doesn't matches the format",
			args: args{
				content: []byte(`
kind: ENHANCEMENTS
body: ": save implicit zookeeper subcluster on hosts changes."
time: 2024-06-21T13:19:18.006693+02:00`,
				),
				unrel: &unreleased{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := handleYamlFile(tt.args.content, tt.args.unrel); (err != nil) != tt.wantErr {
				t.Errorf("handleYamlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
