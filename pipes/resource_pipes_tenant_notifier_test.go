package pipes

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTenantNotifier_Basic(t *testing.T) {
	emailIntegrationHandle := "email.default"
	notifierName := "email-test"
	notifierNameUpdated := "email-test-updated"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantNotifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantNotifierConfig(emailIntegrationHandle, notifierName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantNotifierExists("pipes_tenant_notifier.email"),
					resource.TestCheckResourceAttr("pipes_tenant_notifier.email", "name", notifierName),
				),
			},
			{
				Config: testAccTenantNotifierUpdateConfig(emailIntegrationHandle, notifierNameUpdated),
				Check:  resource.TestCheckResourceAttr("pipes_tenant_notifier.email", "name", notifierNameUpdated),
			},
		},
	})
}

// configs
func testAccTenantNotifierConfig(integrationHandle, notifierName string) string {
	return fmt.Sprintf(`
data "pipes_tenant_integration" "tenant_email_integration" {
	handle = "%s"
}

resource "pipes_tenant_notifier" "email" {
	name        = "%s"
	notifies    = jsonencode([{
		type        = "email"
		integration = data.pipes_tenant_integration.tenant_email_integration.integration_id
		to          = ["user@domain.com"]
	}])
	state       = "enabled"
}`, integrationHandle, notifierName)
}

func testAccTenantNotifierUpdateConfig(integrationHandle, notifierName string) string {
	return fmt.Sprintf(`
data "pipes_tenant_integration" "tenant_email_integration" {
	handle = "%s"
}

resource "pipes_tenant_notifier" "email" {
	name        = "%s"
	notifies    = jsonencode([{
		type        = "email"
		integration = data.pipes_tenant_integration.tenant_email_integration.integration_id
		to          = ["user@domain.com"]
	}])
	state       = "enabled"
}`, integrationHandle, notifierName)
}

// helper functions
func testAccCheckTenantNotifierExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		// Extract tenant handle and user handle from ID
		id := rs.Primary.ID

		client := testAccProvider.Meta().(*PipesClient)
		_, _, err := client.APIClient.TenantNotifiers.Get(context.Background(), id).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckTenantNotifierDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*PipesClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "pipes_tenant_notifier" {
			id := rs.Primary.ID
			_, r, err := client.APIClient.TenantNotifiers.Get(context.Background(), id).Execute()
			if err == nil {
				return fmt.Errorf("tenant notifier still exists")
			}

			// Verify that the error code is 404
			if r.StatusCode != 404 {
				return fmt.Errorf("expected 'not found' error, got %s", err)
			}
		}
	}

	return nil
}
