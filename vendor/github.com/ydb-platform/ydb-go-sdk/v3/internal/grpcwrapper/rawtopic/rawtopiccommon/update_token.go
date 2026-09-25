package rawtopiccommon

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

type UpdateTokenRequest struct {
	Token string
}

func (r *UpdateTokenRequest) ToProto() *Ydb_Topic.UpdateTokenRequest {
	return &Ydb_Topic.UpdateTokenRequest{
		Token: r.Token,
	}
}

type UpdateTokenResponse struct{}

func (r *UpdateTokenResponse) MustFromProto(p *Ydb_Topic.UpdateTokenResponse) {
	// do nothing
}
