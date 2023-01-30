package yandex

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccCMCertificate_managed(t *testing.T) {
	certName := "crt" + acctest.RandString(10)
	certDesc := "Terraform Test"
	folderID := getExampleFolderID()
	managedResource := "yandex_cm_certificate.managed_certificate"
	managedResourceID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexCMCertificateAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create certificate
				Config: testAccCMCertificateManaged(certName, certDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(managedResource, &managedResourceID),
					resource.TestCheckResourceAttr(managedResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(managedResource, "name", certName),
					resource.TestCheckResourceAttr(managedResource, "description", certDesc),
					resource.TestCheckResourceAttr(managedResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(managedResource, "labels.%", "2"),
					resource.TestCheckResourceAttr(managedResource, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(managedResource, "labels.key2", "value2"),
					resource.TestCheckResourceAttr(managedResource, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(managedResource, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(managedResource),
				),
			},
			{
				// Update certificate
				Config: testAccCMCertificateManagedModified(certName+"-modified", certDesc+" edited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(managedResource, nil),
					resource.TestCheckResourceAttrPtr(managedResource, "id", &managedResourceID),
					resource.TestCheckResourceAttr(managedResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(managedResource, "name", certName+"-modified"),
					resource.TestCheckResourceAttr(managedResource, "description", certDesc+" edited"),
					resource.TestCheckResourceAttr(managedResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(managedResource, "labels.%", "3"),
					resource.TestCheckResourceAttr(managedResource, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(managedResource, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(managedResource, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(managedResource, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(managedResource, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(managedResource),
				),
			},
			{
				// Update certificate with recreate
				Config: testAccCMCertificateManagedWildcard(certName+"-wildcard", certDesc+" wildcard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexCMCertificateResourceExists(managedResource, &managedResourceID),
					resource.TestCheckResourceAttr(managedResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(managedResource, "name", certName+"-wildcard"),
					resource.TestCheckResourceAttr(managedResource, "description", certDesc+" wildcard"),
					resource.TestCheckResourceAttr(managedResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(managedResource, "labels.%", "3"),
					resource.TestCheckResourceAttr(managedResource, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(managedResource, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(managedResource, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(managedResource, "status",
						certificatemanager.Certificate_Status_name[int32(certificatemanager.Certificate_VALIDATING)]),
					resource.TestCheckResourceAttr(managedResource, "type",
						certificatemanager.CertificateType_name[int32(certificatemanager.CertificateType_MANAGED)]),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.domain", CMCertificateTestDomainName),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.type", "DNS"),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_name", "_acme-challenge."+CMCertificateTestDomainName+"."),
					resource.TestCheckResourceAttr(managedResource, "challenges.0.dns_type", "CNAME"),
					testAccCheckCreatedAtAttr(managedResource),
				),
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
