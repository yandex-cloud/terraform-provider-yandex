package grpcdebug

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// cleanProtoPayload marshals msg into a JSON object, walks the result tree,
// and replaces any sensitive proto field value with hiddenPlaceholder. Returns
// the resulting tree as a generic interface{} suitable for zap.Any /
// json.Marshal. As a safety net, the antisecret regex sanitizer is applied to
// every remaining string leaf so a token that ends up in an unexpected field
// is still masked.
//
// Lives in sdk-v2 (rather than reusing api/pkg/grpc/util) so sdk-v2 stays free
// of api/* dependencies.
func cleanProtoPayload(msg proto.Message) interface{} {
	if msg == nil {
		return nil
	}
	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
		AllowPartial:  true,
	}.Marshal(msg)
	if err != nil {
		return "ERROR while encoding proto: " + err.Error()
	}
	var tree interface{}
	if err := json.Unmarshal(data, &tree); err != nil {
		return "ERROR while decoding proto JSON: " + err.Error()
	}
	return cleanTree(tree, false)
}

// cleanTree walks the JSON value v. parentSensitive carries information from
// the enclosing key — if v is reached through a sensitive proto field name,
// every leaf below it is replaced with hiddenPlaceholder regardless of its
// shape. Non-sensitive string leaves are still run through sanitizeSecrets so
// regex-detectable secrets (OAuth, IAM tokens, JWTs, …) in unexpected fields
// are masked as well.
func cleanTree(v interface{}, parentSensitive bool) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			_, sensitiveKey := sensitivePayloadFields[k]
			val[k] = cleanTree(child, parentSensitive || sensitiveKey)
		}
		return val
	case []interface{}:
		for i, child := range val {
			val[i] = cleanTree(child, parentSensitive)
		}
		return val
	case string:
		if parentSensitive {
			return hiddenPlaceholder
		}
		return sanitizeSecrets(val)
	default:
		return v
	}
}
