package pipes

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pipes": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	token := os.Getenv("PIPES_TOKEN")
	if token == "" {
		token = os.Getenv("STEAMPIPE_CLOUD_TOKEN")
	}
	if token == "" {
		t.Fatal("`PIPES_TOKEN` or `STEAMPIPE_CLOUD_TOKEN` must be set for acceptance tests.")
	}
}
