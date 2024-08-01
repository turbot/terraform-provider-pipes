package pipes

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccOrganization_Basic(t *testing.T) {
	resourceName := "pipes_organization.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists("pipes_organization.test"),
					resource.TestCheckResourceAttr(
						"pipes_organization.test", "handle", "terraformtest"),
					resource.TestCheckResourceAttr(
						"pipes_organization.test", "display_name", "Terraform Test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationUpdateDisplayNameConfig(),
				Check: resource.TestCheckResourceAttr(
					"pipes_organization.test", "display_name", "Terraform Test Org"),
			},
			{
				Config: testAccOrganizationUpdateHandleConfig(),
				Check: resource.TestCheckResourceAttr(
					"pipes_organization.test", "handle", "terraformtestorg"),
			},
		},
	})
}

// configs
func testAccOrganizationConfig() string {
	return `
resource "pipes_organization" "test" {
	handle       = "terraformtest"
	display_name = "Terraform Test"
}
`
}

func testAccOrganizationUpdateDisplayNameConfig() string {
	return `
resource "pipes_organization" "test" {
	handle       = "terraformtest"
	display_name = "Terraform Test Org"
}
`
}

func testAccOrganizationUpdateHandleConfig() string {
	return `
resource "pipes_organization" "test" {
	handle       = "terraformtestorg"
	display_name = "Terraform Test Org"
}
`
}

// helper functions
func testAccCheckOrganizationExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		client := testAccProvider.Meta().(*PipesClient)
		_, _, err := client.APIClient.Orgs.Get(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckOrganizationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*PipesClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "pipes_organization" {
			_, r, err := client.APIClient.Orgs.Get(context.Background(), rs.Primary.ID).Execute()
			if err == nil {
				return fmt.Errorf("organization still exists")
			}

			// If an organization is deleted, it will return a not found error for subsequent requests
			// for it. If the error is not a 404, then it is unexpected.
			if r.StatusCode != 404 {
				return fmt.Errorf("expected 'not found' error, got %s", err)
			}
		}
	}

	return nil
}
