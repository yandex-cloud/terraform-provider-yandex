package sdkcommon

import (
	"context"
	"fmt"

	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	kmsasymmetricencryption "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	kmsasymmetricsignature "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"

	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func IterateKeys(d *schema.ResourceData, key string) []string {
	size := d.Get(key + ".#").(int)
	var keys []string
	for i := 0; i < size; i++ {
		currentKey := fmt.Sprintf(key+".%d.", i)
		keys = append(keys, currentKey)
	}
	return keys
}

func ExpandStringSlice(v []interface{}) []string {
	if v == nil {
		return nil
	}
	s := make([]string, len(v))
	for i, val := range v {
		s[i] = val.(string)
	}
	return s
}

func ExpandInt64Slice(v []interface{}) []int64 {
	if v == nil {
		return nil
	}
	s := make([]int64, len(v))
	for i, val := range v {
		s[i] = int64(val.(int))
	}
	return s
}

func ExpandLabels(v interface{}) (map[string]string, error) {
	m := make(map[string]string)
	if v == nil {
		return m, nil
	}
	for k, val := range v.(map[string]interface{}) {
		m[k] = val.(string)
	}
	return m, nil
}

func ExpandSecurityGroupIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	sgIdsSet := v.(*schema.Set)
	for _, val := range sgIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
}

func ExpandHostGroupIds(v interface{}) []string {
	if v == nil {
		return nil
	}
	var m []string
	hgIdsSet := v.(*schema.Set)
	for _, val := range hgIdsSet.List() {
		m = append(m, val.(string))
	}
	return m
}

func parseKmsAsymmetricEncryptionAlgorithm(algo string) (kmsasymmetricencryption.AsymmetricEncryptionAlgorithm, error) {
	val, ok := kmsasymmetricencryption.AsymmetricEncryptionAlgorithm_value[algo]
	if !ok {
		return kmsasymmetricencryption.AsymmetricEncryptionAlgorithm(0), fmt.Errorf("value for 'encryption_algorithm' should be one of %s, not `%s`",
			GetJoinedKeys(GetEnumValueMapKeys(kmsasymmetricencryption.AsymmetricEncryptionAlgorithm_value)), algo)
	}
	return kmsasymmetricencryption.AsymmetricEncryptionAlgorithm(val), nil
}

func parseKmsAsymmetricSignatureAlgorithm(algo string) (kmsasymmetricsignature.AsymmetricSignatureAlgorithm, error) {
	val, ok := kmsasymmetricsignature.AsymmetricSignatureAlgorithm_value[algo]
	if !ok {
		return kmsasymmetricsignature.AsymmetricSignatureAlgorithm(0), fmt.Errorf("value for 'signature_algorithm' should be one of %s, not `%s`",
			GetJoinedKeys(GetEnumValueMapKeys(kmsasymmetricsignature.AsymmetricSignatureAlgorithm_value)), algo)
	}
	return kmsasymmetricsignature.AsymmetricSignatureAlgorithm(val), nil
}

//nolint:unused
func securityRuleCidrsFromResourceData(data *schema.ResourceData) (*vpc.CidrBlocks, bool) {
	return SecurityRuleCirds(data.Get("v4_cidr_blocks"), data.Get("v6_cidr_blocks"))
}

func SecurityRuleCirds(v4Blocks, v6Blocks interface{}) (*vpc.CidrBlocks, bool) {
	cidr := new(vpc.CidrBlocks)

	if v4Blocks != nil {
		arr := v4Blocks.([]interface{})
		cidr.V4CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V4CidrBlocks[i] = c.(string)
		}
	}

	if v6Blocks != nil {
		arr := v6Blocks.([]interface{})
		cidr.V6CidrBlocks = make([]string, len(arr))
		for i, c := range arr {
			cidr.V6CidrBlocks[i] = c.(string)
		}
	}

	ok := len(cidr.V4CidrBlocks) > 0 || len(cidr.V6CidrBlocks) > 0
	return cidr, ok
}

//nolint:unused
func securityRulePortsFromResourceData(data *schema.ResourceData) (*vpc.PortRange, error) {
	port := int64(data.Get("port").(int))
	fromPort := int64(data.Get("from_port").(int))
	toPort := int64(data.Get("to_port").(int))

	return SecurityRulePorts(port, fromPort, toPort)
}

func SecurityRulePorts(port, fromPort, toPort int64) (*vpc.PortRange, error) {
	if port == -1 && fromPort == -1 && toPort == -1 {
		return nil, nil
	}

	if port != -1 {
		if fromPort != -1 || toPort != -1 {
			return nil, fmt.Errorf("cannot set from_port/to_port with port")
		}
		fromPort = port
		toPort = port
	} else if fromPort == -1 || toPort == -1 {
		return nil, fmt.Errorf("port or from_port + to_port must be defined")
	}

	return &vpc.PortRange{FromPort: fromPort, ToPort: toPort}, nil
}

func ExpandAndValidateNetworkId(d *schema.ResourceData, config *providerconfig.Config) (string, error) {
	networkID := d.Get("network_id").(string)
	if config.Endpoint == common.DefaultEndpoint && len(networkID) == 0 {
		return "", fmt.Errorf("empty network_id field")
	}
	return networkID, nil
}

func flattenUserSshKey(context context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*providerconfig.Config)
	client := organizationmanagersdk.NewUserSshKeyClient(config.SDK)

	userSSHKeyID := d.Id()
	res, ok := d.GetOk("user_ssh_key_id")
	if ok {
		userSSHKeyID = res.(string)
	}

	userSshKey, err := client.Get(context,
		&organizationmanager.GetUserSshKeyRequest{
			UserSshKeyId: userSSHKeyID,
		})

	if err != nil {
		return HandleNotFoundError(err, d, fmt.Sprintf("User ssh key %q", userSSHKeyID))
	}

	d.SetId(userSSHKeyID)
	_ = d.Set("subject_id", userSshKey.SubjectId)
	_ = d.Set("data", userSshKey.Data)
	_ = d.Set("name", userSshKey.Name)
	_ = d.Set("fingerprint", userSshKey.Fingerprint)
	_ = d.Set("organization_id", userSshKey.OrganizationId)
	_ = d.Set("created_at", GetTimestamp(userSshKey.CreatedAt))
	_ = d.Set("expires_at", GetTimestamp(userSshKey.ExpiresAt))

	return nil
}
