package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func prepareCDNResourceOriginProtocol(d *schema.ResourceData) cdn.OriginProtocol {
	switch d.Get("origin_protocol").(string) {
	case "http":
		return cdn.OriginProtocol_HTTP
	case "https":
		return cdn.OriginProtocol_HTTPS
	case "match":
		return cdn.OriginProtocol_MATCH
	default:
		return cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED
	}
}

func prepareCDNResourceSecondaryHostnames(d *schema.ResourceData) *cdn.SecondaryHostnames {
	hostsSet := d.Get("secondary_hostnames").(*schema.Set)

	var hostNames []string
	for _, hostName := range hostsSet.List() {
		hostNames = append(hostNames, hostName.(string))
	}

	return &cdn.SecondaryHostnames{
		Values: hostNames,
	}
}

func prepareCDNResourceNewSSLCertificate(d *schema.ResourceData) (*cdn.SSLTargetCertificate, error) {
	certSet, ok := d.Get("ssl_certificate").(*schema.Set)
	if !ok || certSet.Len() == 0 {
		return nil, nil
	}

	certFields := certSet.List()[0].(map[string]interface{})

	result := &cdn.SSLTargetCertificate{}

	switch certFields["type"] {
	case cdnSSLCertificateTypeNotUsed:
		result.Type = cdn.SSLCertificateType_DONT_USE
	case cdnSSLCertificateTypeCM:
		result.Type = cdn.SSLCertificateType_CM
	case cdnSSLCertificateTypeLE:
		result.Type = cdn.SSLCertificateType_LETS_ENCRYPT_GCORE
	}

	if result.Type == cdn.SSLCertificateType_CM {
		cmCertID, exist := certFields["certificate_manager_id"]
		if !exist {
			return nil, fmt.Errorf("certificate_manager_id is mandatory field " +
				"for 'certificate_manager' SSL certificate type")
		}
		result.Data = &cdn.SSLCertificateData{
			SslCertificateDataVariant: &cdn.SSLCertificateData_Cm{
				Cm: &cdn.SSLCertificateCMData{
					Id: cmCertID.(string),
				},
			},
		}
	}

	return result, nil
}
