module github.com/terraform-providers/terraform-provider-yandex

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.39
	github.com/c2h5oh/datasize v0.0.0-20171227191756-4eba002a5eae
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/fatih/structs v1.1.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/protobuf v1.3.4
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
	github.com/yandex-cloud/go-genproto v0.0.0-20200226094542-96658959dec4
	github.com/yandex-cloud/go-sdk v0.0.0-20200227103939-01c219f4f3ee
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae // indirect
	google.golang.org/genproto v0.0.0-20200227132054-3f1135a288c9
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
