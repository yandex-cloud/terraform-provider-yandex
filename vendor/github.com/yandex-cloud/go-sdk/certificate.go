package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/certificatemanager"
)

const (
	CertificateServiceID = "certificate-manager"
)

func (sdk *SDK) Certificates() *certificatemanager.CertificateManager {
	return certificatemanager.NewCertificateManager(sdk.getConn(CertificateServiceID))
}
