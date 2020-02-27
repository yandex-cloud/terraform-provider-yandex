module github.com/terraform-providers/terraform-provider-yandex

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.39
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl v0.0.0-20180906183839-65a6292f0157 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/hashicorp/vault v0.10.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/stretchr/testify v1.3.0
	github.com/yandex-cloud/go-genproto v0.0.0-20200217105319-9cedee3a8695
	github.com/yandex-cloud/go-sdk v0.0.0-20200113201139-dc3c759a1204
	golang.org/x/net v0.0.0-20200219183655-46282727080f
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c // indirect
	google.golang.org/genproto v0.0.0-20200218151345-dad8c97a84f5
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
