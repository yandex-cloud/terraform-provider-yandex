module github.com/terraform-providers/terraform-provider-yandex

go 1.12

require (
	github.com/aws/aws-sdk-go v1.19.39
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/fatih/structs v1.1.0
	github.com/golang/protobuf v1.3.5
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/hcl v0.0.0-20180906183839-65a6292f0157 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/hashicorp/vault v0.10.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/stretchr/testify v1.5.1
	github.com/yandex-cloud/go-genproto v0.0.0-20200403155014-561458df7bd3
	github.com/yandex-cloud/go-sdk v0.0.0-20200403155346-96693d292acf
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	google.golang.org/genproto v0.0.0-20200323114720-3f67cca34472
	google.golang.org/grpc v1.28.0
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
