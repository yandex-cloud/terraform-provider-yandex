package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Tag struct {
	Key   string
	Value string
}

func newTag(raw interface{}) *Tag {
	tag := toStringMap(raw)

	var (
		key   string
		value string
		ok    bool
	)
	if key, ok = tag["key"]; !ok {
		return nil
	}
	if value, ok = tag["value"]; !ok {
		return nil
	}
	return &Tag{
		Key:   key,
		Value: value,
	}
}

func newTagsFromS3(tags []*s3.Tag) []Tag {
	out := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, Tag{
			Key:   aws.StringValue(tag.Key),
			Value: aws.StringValue(tag.Value),
		})
	}
	return out
}

func NewTags(raw interface{}) []Tag {
	tags := toStringMap(raw)
	out := make([]Tag, 0, len(tags))
	for k, v := range tags {
		out = append(out, Tag{
			Key:   k,
			Value: v,
		})
	}
	return out
}

func toStringMap(in interface{}) map[string]string {
	if in == nil {
		return nil
	}

	typedValue, ok := in.(map[string]interface{})
	if !ok {
		return nil
	}

	out := make(map[string]string, len(typedValue))

	for k, v := range typedValue {
		value, ok := v.(string)
		if !ok {
			continue
		}

		out[k] = value
	}

	return out
}

func TagsToRaw(tags []Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	out := make(map[string]string, len(tags))
	for _, tag := range tags {
		out[tag.Key] = tag.Value
	}

	return out
}

func TagsToS3(tags []Tag) []*s3.Tag {
	if len(tags) == 0 {
		return nil
	}

	out := make([]*s3.Tag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, &s3.Tag{
			Key:   aws.String(tag.Key),
			Value: aws.String(tag.Value),
		})
	}

	return out
}

func S3TagsToRaw(tags []*s3.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	out := make(map[string]string, len(tags))
	for _, tag := range tags {
		out[*tag.Key] = *tag.Value
	}

	return out
}
