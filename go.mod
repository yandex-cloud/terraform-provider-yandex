module github.com/yandex-cloud/terraform-provider-yandex

go 1.16

require (
	github.com/aws/aws-sdk-go v1.36.30
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/client9/misspell v0.3.4
	github.com/fatih/structs v1.1.0
	github.com/golang/protobuf v1.5.2
	github.com/golangci/golangci-lint v1.43.0
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/hashicorp/vault v0.10.4
	github.com/jen20/awspolicyequivalence v1.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/stretchr/objx v0.1.1
	github.com/stretchr/testify v1.7.0
	github.com/yandex-cloud/go-genproto v0.0.0-20210927112212-0025f65c089d
	github.com/yandex-cloud/go-sdk v0.0.0-20210927113321-18ab1436844a
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
