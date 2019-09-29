module github.com/terraform-providers/terraform-provider-yandex

go 1.12

require (
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/hcl v0.0.0-20180906183839-65a6292f0157 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/hashicorp/vault v0.10.4
	github.com/mitchellh/hashstructure v1.0.0
	github.com/stretchr/testify v1.3.0
	github.com/yandex-cloud/go-genproto v0.0.0-20190928220815-a36c849d0fc1
	github.com/yandex-cloud/go-sdk v0.0.0-20190928221040-2b2c62d945aa
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.23.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
