package pipes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccTenantMember_Basic(t *testing.T) {
	tenantHandle := "terraform" + randomString(3)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantMemberConfig(tenantHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantExists(tenantHandle),
					testAccCheckTenantMemberExists("pipes_tenant_member.test"),
					resource.TestCheckResourceAttr(
						"pipes_tenant_member.test", "role", "member"),
				),
			},
			{
				Config: testAccTenantMemberUpdateConfig(tenantHandle),
				Check: resource.TestCheckResourceAttr(
					"pipes_tenant_member.test", "role", "owner"),
			},
		},
	})
}

// configs
func testAccTenantMemberConfig(tenantHandle string) string {
	return fmt.Sprintf(`
provider "pipes" {}

data "pipes_tenant" "test_tenant" {
	handle = "%s"
}

# Please provide a valid email
resource "pipes_tenant_member" "test" {
	tenant = pipes_tenant.test_tenant.handle
	email        = "das.siddhartha992@gmail.com" # "user@domain.com"
	role         = "member"
}`, tenantHandle)
}

func testAccTenantMemberUpdateConfig(tenantHandle string) string {
	return fmt.Sprintf(`
provider "pipes" {}

data "pipes_tenant" "test_tenant" {
	handle = "%s"
}

# Please provide a valid email
resource "pipes_tenant_member" "test" {
	tenant = pipes_tenant.test_tenant.handle
	email        = "das.siddhartha992@gmail.com" # "user@domain.com"
	role         = "owner"
}`, tenantHandle)
}

// helper functions
func testAccCheckTenantMemberExists(resource string) resource.TestCheckFunc {
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
		idParts := strings.Split(id, "/")
		if len(idParts) < 2 {
			return fmt.Errorf("unexpected format of ID (%q), expected <tenant_handle>/<user_handle>", id)
		}

		client := testAccProvider.Meta().(*PipesClient)
		_, _, err := client.APIClient.TenantMembers.Get(context.Background(), idParts[0], idParts[1]).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckTenantMemberDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*PipesClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "pipes_tenant_member" {
			// Extract tenant handle and user handle from ID
			id := rs.Primary.ID
			idParts := strings.Split(id, "/")
			if len(idParts) < 2 {
				return fmt.Errorf("unexpected format of ID (%q), expected <tenant_handle>/<user_handle>", id)
			}

			_, r, err := client.APIClient.TenantMembers.Get(context.Background(), idParts[0], idParts[1]).Execute()
			if err == nil {
				return fmt.Errorf("tenant member still exists")
			}

			// If a tenant is deleted, all the members will lost access to that tenant
			// If anyone try to get that deleted resource, it will always return `403 Forbidden` error
			if r.StatusCode != 403 {
				return fmt.Errorf("expected 'forbidden' error, got %s", err)
			}
		}
	}

	return nil
}

func testAccCheckTenantExists(tenantHandle string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)
		ctx := context.Background()
		var err error
		var r *http.Response

		// check if tenant  is created
		_, r, err = client.APIClient.Tenants.Get(ctx, tenantHandle).Execute()
		if err != nil {
			if r.StatusCode != 403 {
				return fmt.Errorf("error fetching tenant with handle %s. %s", tenantHandle, err)
			}
		}
		return nil
	}
}
