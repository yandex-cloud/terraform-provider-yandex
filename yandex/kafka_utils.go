package yandex

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

//go:generate ../scripts/mockgen.sh KafkaTopicModifier

type KafkaTopicModifier interface {
	CreateKafkaTopic(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error
	DeleteKafkaTopic(ctx context.Context, d *schema.ResourceData, topicName string) error
	UpdateKafkaTopic(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec, paths []string) error
}

type KafkaTopicManager struct {
	Config *Config
}

func NewKafkaTopicManager(config *Config) *KafkaTopicManager {
	return &KafkaTopicManager{Config: config}
}

func (tm *KafkaTopicManager) CreateKafkaTopic(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error {
	return createKafkaTopic(ctx, tm.Config, d, topicSpec)
}

func (tm *KafkaTopicManager) DeleteKafkaTopic(ctx context.Context, d *schema.ResourceData, topicName string) error {
	return deleteKafkaTopic(ctx, tm.Config, d, topicName)
}

func (tm *KafkaTopicManager) UpdateKafkaTopic(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec, paths []string) error {
	return updateKafkaTopic(ctx, tm.Config, d, topicSpec, paths)
}
