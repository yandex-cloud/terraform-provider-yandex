module github.com/yandex-cloud/terraform-provider-yandex

go 1.16

require (
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.37.0
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/client9/misspell v0.3.4
	github.com/fatih/structs v1.1.0
	github.com/frankban/quicktest v1.14.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
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
	github.com/hashicorp/hcl/v2 v2.8.2 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/hashicorp/vault v0.10.4
	github.com/jen20/awspolicyequivalence v1.1.0
	github.com/keybase/go-crypto v0.0.0-20200123153347-de78d2cb44f4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/stretchr/objx v0.1.1
	github.com/stretchr/testify v1.7.0
	github.com/yandex-cloud/go-genproto v0.0.0-20220307143823-ae6fd1037836
	github.com/yandex-cloud/go-sdk v0.0.0-20220307144046-5eb2045b0e5f
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect
	google.golang.org/genproto v0.0.0-20220310185008-1973136f34c6
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.27.1
)

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/spf13/afero => github.com/spf13/afero v1.2.2
)
