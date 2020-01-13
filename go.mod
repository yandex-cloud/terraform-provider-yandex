module github.com/terraform-providers/terraform-provider-yandex

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.39
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.3.2
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
	github.com/yandex-cloud/go-genproto v0.0.0-20200113111713-27a750dfd05a
	github.com/yandex-cloud/go-sdk v0.0.0-20200113201139-dc3c759a1204
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/sys v0.0.0-20200113162924-86b910548bc1 // indirect
	google.golang.org/genproto v0.0.0-20200113173426-e1de0a7b01eb
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
