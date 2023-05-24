module github.com/yandex-cloud/terraform-provider-yandex

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.0
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/client9/misspell v0.3.4
	github.com/fatih/structs v1.1.0
	github.com/frankban/quicktest v1.14.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/golangci/golangci-lint v1.43.0
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.14.0
	github.com/hashicorp/vault v0.10.4
	github.com/jen20/awspolicyequivalence v1.1.0
	github.com/keybase/go-crypto v0.0.0-20200123153347-de78d2cb44f4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/stretchr/objx v0.1.1
	github.com/stretchr/testify v1.7.1
	github.com/yandex-cloud/go-genproto v0.0.0-20230502091605-c1556b468ba3
	github.com/yandex-cloud/go-sdk v0.0.0-20230502092042-98f99e999085
	github.com/ydb-platform/terraform-provider-ydb v0.0.10
	golang.org/x/net v0.8.0
	golang.org/x/tools v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20221107162902-2d387536bcdd
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/spf13/afero => github.com/spf13/afero v1.2.2
)
