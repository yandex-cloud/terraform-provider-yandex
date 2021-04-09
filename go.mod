module github.com/yandex-cloud/terraform-provider-yandex

go 1.16

require (
	github.com/aws/aws-sdk-go v1.36.30
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/fatih/structs v1.1.0
	github.com/golang/protobuf v1.4.2
	github.com/golangci/golangci-lint v1.39.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/go-getter v1.4.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/hashicorp/vault v0.10.4
	github.com/jen20/awspolicyequivalence v1.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/yandex-cloud/go-genproto v0.0.0-20210326132454-24349c492ce9
	github.com/yandex-cloud/go-sdk v0.0.0-20210326140609-dcebefcc0553
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	google.golang.org/genproto v0.0.0-20200707001353-8e8330bf89df
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
