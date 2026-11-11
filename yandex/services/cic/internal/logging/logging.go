package logging

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var bgpMD5Field = regexp.MustCompile(`"(?:peer_bgp_md5_key|peerBgpMd5Key)"\s*:\s*"(?:\\.|[^"\\])*"`)

func MaskContext(ctx context.Context) context.Context {
	ctx = tflog.MaskMessageRegexes(ctx, bgpMD5Field)
	return tflog.OmitLogWithMessageStrings(ctx, "**ERROR json mashal failed:")
}
