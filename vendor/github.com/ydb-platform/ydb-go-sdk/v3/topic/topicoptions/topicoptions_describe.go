package topicoptions

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawtopic"

// DescribeOption type for options of describe method.
type DescribeOption func(req *rawtopic.DescribeTopicRequest)

// IncludePartitionStats additionally fetches DescribeTopicResult.PartitionInfo.PartitioningSettings from server
func IncludePartitionStats() DescribeOption {
	return func(req *rawtopic.DescribeTopicRequest) {
		req.IncludeStats = true
	}
}

// DescribeConsumerOption type for options of describe consumer method.
type DescribeConsumerOption func(req *rawtopic.DescribeConsumerRequest)

// IncludeConsumerStats additionally fetches
// DescribeConsumerResult.DescribeConsumerResultPartitionInfo.PartitionConsumerStats from server
func IncludeConsumerStats() DescribeConsumerOption {
	return func(req *rawtopic.DescribeConsumerRequest) {
		req.IncludeStats = true
	}
}
