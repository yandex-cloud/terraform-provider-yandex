package yandex

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
)

func init() {
	resource.AddTestSweepers("yandex_cm_certificate", &resource.Sweeper{
		Name: "yandex_cm_certificate",
		F:    testSweepCMCertificate,
	})
}

func generateRandomDomainName() string {
	adjectives := [...]string{"additional", "afraid", "angry", "anxious", "asleep", "attentive", "available", "basic", "beautiful", "big", "boring", "brave", "bright", "busy", "calm", "careful", "cheap", "clean", "clever", "cold", "comfortable", "confident", "conscious", "constant", "convenient", "cool", "correct", "curious", "dangerous", "dark", "deep", "different", "difficult", "dirty", "easy", "efficient", "empty", "every", "exact", "exciting", "expensive", "fair", "famous", "fast", "fat", "fine", "firm", "flat", "foreign", "formal", "former", "free", "fresh", "friendly", "frightful", "full", "funny", "gorgeous", "guilty", "happy", "hard", "healthy", "heavy", "helpful", "historical", "honest", "hot", "huge", "hungry", "ill", "illegal", "important", "impossible", "independent", "informal", "innocent", "interesting", "international", "kind", "large", "leading", "legal", "light", "little", "lonely", "long", "loose", "loud", "lucky", "necessary", "nice", "normal", "obvious", "official", "old", "opposite", "perfect", "pleasant", "polite", "poor", "popular", "possible", "powerful", "quiet", "rare", "recent", "relevant", "remarkable", "remote", "responsible", "rich", "rude", "sad", "safe", "secure", "sensible", "short", "silly", "similar", "slow", "small", "smooth", "strange", "strict", "strong", "successful", "sudden", "suitable", "suspicious", "sweet", "tall", "tasty", "terrible", "thin", "thirsty", "tight", "tiny", "tired", "traditional", "typical", "useful", "usual", "valuable", "warm", "weak", "weird", "wide", "wise", "wonderful", "young"}
	nouns := [...]string{"action", "activity", "age", "air", "animal", "area", "authority", "bank", "body", "book", "building", "business", "car", "case", "centre", "century", "change", "child", "city", "community", "company", "condition", "control", "country", "course", "court", "day", "decision", "development", "door", "education", "effect", "end", "example", "experience", "eye", "face", "fact", "family", "father", "field", "figure", "flat", "food", "form", "friend", "game", "girl", "government", "group", "guy", "hand", "head", "health", "history", "home", "hour", "house", "idea", "industry", "information", "interest", "job", "kid", "kind", "language", "law", "level", "life", "line", "love", "man", "manager", "manner", "market", "million", "mind", "minute", "moment", "money", "month", "morning", "mother", "name", "need", "night", "number", "office", "opportunity", "order", "paper", "parent", "part", "party", "people", "period", "person", "place", "plan", "point", "police", "policy", "position", "power", "president", "price", "problem", "process", "programme", "project", "quality", "question", "reason", "relationship", "report", "research", "rest", "result", "road", "room", "school", "sense", "service", "side", "society", "staff", "story", "street", "student", "study", "support", "system", "table", "teacher", "team", "term", "thing", "time", "type", "use", "view", "war", "water", "way", "week", "woman", "word", "work", "world", "year"}

	var buffer bytes.Buffer
	buffer.WriteString(adjectives[rand.Intn(len(adjectives))])
	if rand.Intn(2) == 0 {
		if rand.Intn(2) == 0 {
			buffer.WriteString("-")
		}
		buffer.WriteString(adjectives[rand.Intn(len(adjectives))])
	}
	if rand.Intn(2) == 0 {
		buffer.WriteString("-")
	}
	buffer.WriteString(nouns[rand.Intn(len(nouns))])
	if rand.Intn(2) == 0 {
		if rand.Intn(2) == 0 {
			buffer.WriteString("-")
		}
		buffer.WriteString(nouns[rand.Intn(len(nouns))])
	}
	if rand.Intn(2) == 0 {
		if rand.Intn(2) == 0 {
			buffer.WriteString("-")
		}
		buffer.WriteString(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	}
	buffer.WriteString(".ru")
	return buffer.String()
}

var CMCertificateTestDomainName = generateRandomDomainName()
var CMCertificateTestSelfSignedCertificate = "-----BEGIN CERTIFICATE-----\nMIICqjCCAZICCQCdETTBqCSthjANBgkqhkiG9w0BAQUFADAWMRQwEgYDVQQDDAtl\neGFtcGxlLmNvbTAgFw0yMzA0MjMwOTQ4MTNaGA83NDk5MDIxMzA5NDgxM1owFjEU\nMBIGA1UEAwwLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQC23EElFIrqw/mAmBI8woU9YdJnscazH1GWbYo59ReU65kKsaHT4qm+J65H\nZdUjzmx/pExYqowFCZz3s+GDa2xiGN47sTKTPz+rUYkLdLpZoJSIj9AwbMfF6BJt\nFLZ3A2hPLGa1V64Au1974mlaCaakFJqxdf1j4OMQbyxlqM9xs8sGCFK59oJT1phI\nLxqEuTkvO1DTeBxHrrsl3PyTMcnp+aatUjxaAhUXURYfi3P2G2l/2TJUBNkvc1T7\nXHBGEgNlgoZJrP6X3H3IFl8/6l0HnEXiZdaTargasnkThZUHflUmotjdLl+7mZ8M\n/ktenIBYkQOq3k/EwTOHvdglmQBJAgMBAAEwDQYJKoZIhvcNAQEFBQADggEBADx6\ndGs/S8MMfa34vN7WLIn6R7/l4RWDVEJ8CHpQRwq5PaHamuYsEsT7A1N+nFEuTqw6\nUFrjkMhENGTxJl0SdezU0RePmouXGwNRyG2eC1PXo14e30xTbBctVNI+Ntj2H+lt\nGsyBHISBtAIarvZgv4HsRGw1OSDwunBFQD/lAQhlAg1yCSMk/oy5wjgrCLUJTm6j\nV0xhdCub4wZw+gfug1Y5XPLED1r3ne34BSpOatIS3sqjsexw6133Os2XgIXjO1IN\nFtG3EgAc/EIJAXVfbzT8azaHfjD4pZdO0RAwr8sQHOQqI/MzJCo11lV/rd5CNfpc\niv78dk8SGtlMtunFQk8=\n-----END CERTIFICATE-----\n"
var CMCertificateTestPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAttxBJRSK6sP5gJgSPMKFPWHSZ7HGsx9Rlm2KOfUXlOuZCrGh\n0+KpvieuR2XVI85sf6RMWKqMBQmc97Phg2tsYhjeO7Eykz8/q1GJC3S6WaCUiI/Q\nMGzHxegSbRS2dwNoTyxmtVeuALtfe+JpWgmmpBSasXX9Y+DjEG8sZajPcbPLBghS\nufaCU9aYSC8ahLk5LztQ03gcR667Jdz8kzHJ6fmmrVI8WgIVF1EWH4tz9htpf9ky\nVATZL3NU+1xwRhIDZYKGSaz+l9x9yBZfP+pdB5xF4mXWk2q4GrJ5E4WVB35VJqLY\n3S5fu5mfDP5LXpyAWJEDqt5PxMEzh73YJZkASQIDAQABAoIBADcWLS3dfWfx99Ts\nevoA46C1OmxwmtpVQf/eKfkBw8PiIa2eC5FIRDh7vb3WiJoL0pW1Siaf4iSWW8on\nT3WGxBTdRv2WiRTgxe53VqCz3nunq3dkU6Ry8M/G9N4Vkk5SIXdQefSBYHLp/37T\nm0c7hw8BAgUZ9WbEVcMaqrZJX4zxyWxDIUoO7KUDU/VW0thi1iw3+bILc4wn/zgI\nxiZFZl/bvPdri0U0dkoUWk6ZyiC7czmrqb82t4vrjN3NR6obZbfeCtXlAp51stYV\neH7ciXk3HChEOAQ5BT3DyQjhqgB/HrDuEbiIeeLOGMATyOqMXy3T5kxbxQ9QtXyF\nc3NljQECgYEA4XTSUaQadRPTs1upnOvm0FPv8gkkCYhV5DxOqotM9AO3PRAFg1NG\nLrnOXB409W6gxSkR2ore0oYP28bdaDnM/O1Msjqz86tcOLjULpXFRZG1SuNkOC90\nBDyL5J9cwaTdZqXSHQooljzxsRCIy6c1F3X9swSljthxYjXcmhRzHWkCgYEAz6Ij\nYIRsVc/jqHOwNpbKT6lo727IO8dnu7iOH6EFxn3jvtrfHrNp+Ghk7+8bBRcNDHJI\nRw914/sNFGKQkbj1UlwaC3Pk33dyftDVUJjcJmoOYZVI2olzH9FTjP1CbGBYTZza\nMN2+UpZR/h1IlSVbbp0cu9CpGpyzIa8ZAyK7D+ECgYEA1JjbZp7vT+11WJEb/Nw6\nV8J+5eYWtGJ6U/FGYO1gkE0cshj0ieSxrogJfrYBTFqYgbJ7onAHM8+1DpKU3555\nnRuLkhlm7WRuXxJzCsayMir3IHoSXCTrKr+JTvmzduqm2A+PdVDJ+vnXExe7VwcC\nOnBJ3lCIaY3SRUDzF9wmvNkCgYAWg1wWoQUmIM5se27F3H+/N307SOXJJYvn3ND8\nOPdpWEkTbqP2rjl1R8x5/5EMcj1l9hZELjb4K0Z1yWInish+z6G7UCum10rA2V/n\nx0tHlwRMLGWj3HdxIb9PcD59hczNTY6S8dgrGEV3qjEuishpK/vrmWpcilUZ9+Rc\nZK2nwQKBgCOAK4aJLPbdzJAm932i55UizfZBtbMRC8d99PucnT4APzY0lQEoFn+U\nGpuQWg04DIe/gpvKXFqLPwZ0RYJiypKDVllQts9SwuvBjsf80ZzfaT4Nrghn7Nq/\nbmaNnUm58lOvMyaSIsfw0B8uKOh+YU+kJcsdO9z7dY18BWHEAlMH\n-----END RSA PRIVATE KEY-----\n"

func TestAccCMCertificate_managed(t *testing.T) {
	certName := "crt" + acctest.RandString(10)
	certDesc := "Terraform Test"
	folderID := getExampleFolderID()
	resourceName := "yandex_cm_certificate.managed_certificate"
	resourceID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexCMCertificateAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create certificate
				Config: testAccCMCertificateManaged(certName, certDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(resourceName, &resourceID),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", certName),
					resource.TestCheckResourceAttr(resourceName, "description", certDesc),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(resourceName, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
			{
				// Update certificate
				Config: testAccCMCertificateManagedModified(certName+"-modified", certDesc+" edited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(resourceName, nil),
					resource.TestCheckResourceAttrPtr(resourceName, "id", &resourceID),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", certName+"-modified"),
					resource.TestCheckResourceAttr(resourceName, "description", certDesc+" edited"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(resourceName, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(resourceName, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(resourceName, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(resourceName, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
			{
				ResourceName:      "yandex_cm_certificate.managed_certificate",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"managed", // certificate contents is not returned
				},
			},
			{
				// Update certificate with recreate
				Config: testAccCMCertificateManagedWildcard(certName+"-wildcard", certDesc+" wildcard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(resourceName, &resourceID),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", certName+"-wildcard"),
					resource.TestCheckResourceAttr(resourceName, "description", certDesc+" wildcard"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(resourceName, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(resourceName, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(resourceName, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(resourceName, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(resourceName, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
		},
	})
}

func TestAccCMCertificate_selfManaged(t *testing.T) {
	certName := "crt" + acctest.RandString(10) + "-self-managed"
	certDesc := "Terraform Test Self Managed Certificate"
	folderID := getExampleFolderID()
	resourceName := "yandex_cm_certificate.self_managed_certificate"
	resourceID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexCMCertificateAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Upload self-managed certificate
				Config: testAccCMCertificateSelfManaged(certName, certDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(resourceName, &resourceID),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", certName),
					resource.TestCheckResourceAttr(resourceName, "description", certDesc),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "serial", "9d1134c1a824ad86"),
					resource.TestCheckResourceAttr(resourceName, "challenges.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "not_after", "7499-02-13T09:48:13Z"),
					resource.TestCheckResourceAttr(resourceName, "not_before", "2023-04-23T09:48:13Z"),
					resource.TestCheckResourceAttr(resourceName, "managed.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_managed.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_managed.0.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "self_managed.0.certificate", CMCertificateTestSelfSignedCertificate),
					resource.TestCheckResourceAttr(resourceName, "self_managed.0.private_key", CMCertificateTestPrivateKey),
					resource.TestCheckResourceAttr(resourceName, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_ISSUED)]),
					resource.TestCheckResourceAttr(resourceName, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_IMPORTED)]),
					testAccCheckCreatedAtAttr(resourceName),
				),
			},
			{
				ResourceName:      "yandex_cm_certificate.self_managed_certificate",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"self_managed", // certificate contents is not returned
				},
			},
		},
	})
}

func testAccCMCertificateManaged(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_cm_certificate" "managed_certificate" {
  name        = "%v"
  description = "%v"
  labels      = {
    key1 = "value1"
    key2 = "value2"
  }
  deletion_protection = true
  domains = ["%v"]
  managed {
    challenge_type = "DNS_CNAME"
    challenge_count = 1
  }
}
`, name, desc, CMCertificateTestDomainName)
}

func testAccCMCertificateManagedModified(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_cm_certificate" "managed_certificate" {
  name                = "%v"
  description         = "%v"
  labels              = {
    key1 = "value10"
    key3 = "value30"
    key4 = "value40"
  }
  deletion_protection = false
  domains = ["%v"]
  managed {
    challenge_type = "DNS_CNAME"
  }
}
`, name, desc, CMCertificateTestDomainName)
}

func testAccCMCertificateManagedWildcard(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_cm_certificate" "managed_certificate" {
  name                = "%v"
  description         = "%v"
  labels              = {
    key1 = "value10"
    key3 = "value30"
    key4 = "value40"
  }
  deletion_protection = false
  domains = ["%v", "*.%[3]v"]
  managed {
    challenge_type = "DNS_CNAME"
    challenge_count = 1
  }
}
`, name, desc, CMCertificateTestDomainName)
}

func testAccCMCertificateSelfManaged(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_cm_certificate" "self_managed_certificate" {
 name        = "%v"
 description = "%v"
 labels      = {
   key1 = "value1"
   key2 = "value2"
 }
 deletion_protection = false
 self_managed {
   certificate = <<EOF
%vEOF
   private_key = <<EOF
%vEOF
 }
}
`,
		name,
		desc,
		CMCertificateTestSelfSignedCertificate,
		CMCertificateTestPrivateKey,
	)
}

// If idPtr is provided:
// - idPtr must be different from the resource ID (you can use it to check that the resource is different)
// - resource ID will be set to idPtr
func testAccCheckYandexCMCertificateResourceExists(r string, idPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found resource: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for the resource: %s", r)
		}
		if idPtr != nil {
			if rs.Primary.ID == *idPtr {
				return fmt.Errorf("ID %s of resource %s is the same", rs.Primary.ID, r)
			}
			*idPtr = rs.Primary.ID
		}
		return nil
	}
}

func testAccCheckYandexCMCertificateAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cm_certificate" {
			continue
		}
		if err := testAccCheckYandexCMCertificateDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexCMCertificateDestroyed(id string) error {
	config := testAccProvider.Meta().(*Config)
	_, err := config.sdk.Certificates().Certificate().Get(context.Background(), &certificatemanager.GetCertificateRequest{
		CertificateId: id,
	})
	if err == nil {
		return fmt.Errorf("CMCertificate %s still exists", id)
	}
	return nil
}

func testSweepCMCertificate(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &certificatemanager.ListCertificatesRequest{FolderId: conf.FolderID}
	it := conf.sdk.Certificates().Certificate().CertificateIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		certificate := it.Value()

		if len(certificate.Labels) > 0 {
			if certificate.Labels["sweeper-skip-deletion"] == "1" {
				continue
			}
		}

		id := certificate.GetId()
		if !sweepCMCertificate(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep certificate %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCMCertificate(conf *Config, id string) bool {
	return sweepWithRetry(sweepCMCertificateOnce, conf, "Certificate", id)
}

func sweepCMCertificateOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexCMCertificateDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Certificates().Certificate().Delete(ctx, &certificatemanager.DeleteCertificateRequest{
		CertificateId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
