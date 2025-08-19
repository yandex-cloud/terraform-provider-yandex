package yandex_cm_certificate_iam_member_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
)

var CMCertificateTestSelfSignedCertificate = "-----BEGIN CERTIFICATE-----\nMIICqjCCAZICCQCdETTBqCSthjANBgkqhkiG9w0BAQUFADAWMRQwEgYDVQQDDAtl\neGFtcGxlLmNvbTAgFw0yMzA0MjMwOTQ4MTNaGA83NDk5MDIxMzA5NDgxM1owFjEU\nMBIGA1UEAwwLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQC23EElFIrqw/mAmBI8woU9YdJnscazH1GWbYo59ReU65kKsaHT4qm+J65H\nZdUjzmx/pExYqowFCZz3s+GDa2xiGN47sTKTPz+rUYkLdLpZoJSIj9AwbMfF6BJt\nFLZ3A2hPLGa1V64Au1974mlaCaakFJqxdf1j4OMQbyxlqM9xs8sGCFK59oJT1phI\nLxqEuTkvO1DTeBxHrrsl3PyTMcnp+aatUjxaAhUXURYfi3P2G2l/2TJUBNkvc1T7\nXHBGEgNlgoZJrP6X3H3IFl8/6l0HnEXiZdaTargasnkThZUHflUmotjdLl+7mZ8M\n/ktenIBYkQOq3k/EwTOHvdglmQBJAgMBAAEwDQYJKoZIhvcNAQEFBQADggEBADx6\ndGs/S8MMfa34vN7WLIn6R7/l4RWDVEJ8CHpQRwq5PaHamuYsEsT7A1N+nFEuTqw6\nUFrjkMhENGTxJl0SdezU0RePmouXGwNRyG2eC1PXo14e30xTbBctVNI+Ntj2H+lt\nGsyBHISBtAIarvZgv4HsRGw1OSDwunBFQD/lAQhlAg1yCSMk/oy5wjgrCLUJTm6j\nV0xhdCub4wZw+gfug1Y5XPLED1r3ne34BSpOatIS3sqjsexw6133Os2XgIXjO1IN\nFtG3EgAc/EIJAXVfbzT8azaHfjD4pZdO0RAwr8sQHOQqI/MzJCo11lV/rd5CNfpc\niv78dk8SGtlMtunFQk8=\n-----END CERTIFICATE-----\n"
var CMCertificateTestPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAttxBJRSK6sP5gJgSPMKFPWHSZ7HGsx9Rlm2KOfUXlOuZCrGh\n0+KpvieuR2XVI85sf6RMWKqMBQmc97Phg2tsYhjeO7Eykz8/q1GJC3S6WaCUiI/Q\nMGzHxegSbRS2dwNoTyxmtVeuALtfe+JpWgmmpBSasXX9Y+DjEG8sZajPcbPLBghS\nufaCU9aYSC8ahLk5LztQ03gcR667Jdz8kzHJ6fmmrVI8WgIVF1EWH4tz9htpf9ky\nVATZL3NU+1xwRhIDZYKGSaz+l9x9yBZfP+pdB5xF4mXWk2q4GrJ5E4WVB35VJqLY\n3S5fu5mfDP5LXpyAWJEDqt5PxMEzh73YJZkASQIDAQABAoIBADcWLS3dfWfx99Ts\nevoA46C1OmxwmtpVQf/eKfkBw8PiIa2eC5FIRDh7vb3WiJoL0pW1Siaf4iSWW8on\nT3WGxBTdRv2WiRTgxe53VqCz3nunq3dkU6Ry8M/G9N4Vkk5SIXdQefSBYHLp/37T\nm0c7hw8BAgUZ9WbEVcMaqrZJX4zxyWxDIUoO7KUDU/VW0thi1iw3+bILc4wn/zgI\nxiZFZl/bvPdri0U0dkoUWk6ZyiC7czmrqb82t4vrjN3NR6obZbfeCtXlAp51stYV\neH7ciXk3HChEOAQ5BT3DyQjhqgB/HrDuEbiIeeLOGMATyOqMXy3T5kxbxQ9QtXyF\nc3NljQECgYEA4XTSUaQadRPTs1upnOvm0FPv8gkkCYhV5DxOqotM9AO3PRAFg1NG\nLrnOXB409W6gxSkR2ore0oYP28bdaDnM/O1Msjqz86tcOLjULpXFRZG1SuNkOC90\nBDyL5J9cwaTdZqXSHQooljzxsRCIy6c1F3X9swSljthxYjXcmhRzHWkCgYEAz6Ij\nYIRsVc/jqHOwNpbKT6lo727IO8dnu7iOH6EFxn3jvtrfHrNp+Ghk7+8bBRcNDHJI\nRw914/sNFGKQkbj1UlwaC3Pk33dyftDVUJjcJmoOYZVI2olzH9FTjP1CbGBYTZza\nMN2+UpZR/h1IlSVbbp0cu9CpGpyzIa8ZAyK7D+ECgYEA1JjbZp7vT+11WJEb/Nw6\nV8J+5eYWtGJ6U/FGYO1gkE0cshj0ieSxrogJfrYBTFqYgbJ7onAHM8+1DpKU3555\nnRuLkhlm7WRuXxJzCsayMir3IHoSXCTrKr+JTvmzduqm2A+PdVDJ+vnXExe7VwcC\nOnBJ3lCIaY3SRUDzF9wmvNkCgYAWg1wWoQUmIM5se27F3H+/N307SOXJJYvn3ND8\nOPdpWEkTbqP2rjl1R8x5/5EMcj1l9hZELjb4K0Z1yWInish+z6G7UCum10rA2V/n\nx0tHlwRMLGWj3HdxIb9PcD59hczNTY6S8dgrGEV3qjEuishpK/vrmWpcilUZ9+Rc\nZK2nwQKBgCOAK4aJLPbdzJAm932i55UizfZBtbMRC8d99PucnT4APzY0lQEoFn+U\nGpuQWg04DIe/gpvKXFqLPwZ0RYJiypKDVllQts9SwuvBjsf80ZzfaT4Nrghn7Nq/\nbmaNnUm58lOvMyaSIsfw0B8uKOh+YU+kJcsdO9z7dY18BWHEAlMH\n-----END RSA PRIVATE KEY-----\n"

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCMCertificateIamMember_basic(t *testing.T) {
	var certificate certificatemanager.Certificate
	name := acctest.RandomWithPrefix("tf-cm-certificate")
	certificateContent, privateKey, err := getSelfSignedCertificate()
	if err != nil {
		t.Fatal(err)
	}

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					test.TestAccCheckCMEmptyIam("yandex_cm_certificate.test"),
				),
			},
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, role, userID),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					test.TestAccCheckCMCertificateIam("yandex_cm_certificate.test", role, []string{userID}),
				),
			},
			{
				ResourceName: "yandex_cm_certificate_iam_member.test-member",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return certificate.Id + " " + role + " " + userID, nil
				},
				ImportState: true,
			},
			{
				Config: testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, "", ""),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccCheckCMCertificateExists("yandex_cm_certificate.test", &certificate),
					test.TestAccCheckCMEmptyIam("yandex_cm_certificate.test"),
				),
			},
		},
	})
}

func testAccCMCertificateIamMemberBasic(name, certificateContent, privateKey, role, member string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(`
resource "yandex_cm_certificate" "test" {
  name = "%s"
  self_managed {
	certificate = <<EOF
%s
EOF
	private_key = <<EOF
%s
EOF
  }
}
		`, name, certificateContent, privateKey))

	if role != "" && member != "" {
		builder.WriteString(fmt.Sprintf(`
resource "yandex_cm_certificate_iam_member" "test-member" {
  certificate_id = yandex_cm_certificate.test.id
  role   = "%s"
  member = "%s"
}
		`, role, member))
	}
	return builder.String()
}

func getSelfSignedCertificate() (string, string, error) {
	return CMCertificateTestSelfSignedCertificate, CMCertificateTestPrivateKey, nil
}
