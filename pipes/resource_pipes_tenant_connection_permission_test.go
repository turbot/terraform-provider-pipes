package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites

func TestAccTenantConnectionPermission_Basic(t *testing.T) {
	orgResource1 := "pipes_organization.test_org_1"
	orgResource2 := "pipes_organization.test_org_2"
	connResource := "pipes_tenant_connection.connection_1"
	permissionResource := "pipes_tenant_connection_permission.permission_1"
	orgHandle1 := "org" + randomString(5)
	orgHandle2 := "org" + randomString(6)
	connHandle := "aws" + randomString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantConnectionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConnectionPermissionConfig(orgHandle1, orgHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource1),
					testAccCheckOrganizationExists(orgResource2),
					testAccCheckTenantConnectionExists(connResource),
					testAccCheckTenantConnectionPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle1),
					testAccCheckTenantConnectionAccess(connResource, orgHandle1, orgHandle2),
				),
			},
			{
				ResourceName: permissionResource,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccTenantConnectionPermissionUpdateConfig(orgHandle1, orgHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource1),
					testAccCheckOrganizationExists(orgResource2),
					testAccCheckTenantConnectionExists(connResource),
					testAccCheckTenantConnectionPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle2),
					testAccCheckTenantConnectionAccess(connResource, orgHandle2, orgHandle1),
				),
			},
		},
	})
}

// configs
func testAccTenantConnectionPermissionConfig(orgHandle1, orgHandle2, connHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org_1" {
	handle       = "%s"
	display_name = "Terraform Test Org 1"
}

resource "pipes_organization" "test_org_2" {
	handle       = "%s"
	display_name = "Terraform Test Org 2"
}

resource "pipes_tenant_connection" "connection_1" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}

resource "pipes_tenant_connection_permission" "permission_1" {
	connection_handle = pipes_tenant_connection.connection_1.handle
	identity_handle   = pipes_organization.test_org_1.handle
}`, orgHandle1, orgHandle2, connHandle)
}

func testAccTenantConnectionPermissionUpdateConfig(orgHandle1, orgHandle2, connHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org_1" {
	handle       = "%s"
	display_name = "Terraform Test Org 1"
}

resource "pipes_organization" "test_org_2" {
	handle       = "%s"
	display_name = "Terraform Test Org 2"
}

resource "pipes_tenant_connection" "connection_1" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}

resource "pipes_tenant_connection_permission" "permission_1" {
	connection_handle = pipes_tenant_connection.connection_1.handle
	identity_handle   = pipes_organization.test_org_2.handle
}`, orgHandle1, orgHandle2, connHandle)
}

// testAccCheckTenantConnectionPermissionDestroy verifies the connection permission has been destroyed
func testAccCheckTenantConnectionPermissionDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_tenant_connection_permission" {
			continue
		}

		connectionHandle := rs.Primary.Attributes["connection_handle"]
		permissionId := rs.Primary.Attributes["permission_id"]

		_, r, err = client.APIClient.TenantConnections.GetPermission(ctx, connectionHandle, permissionId).Execute()
		if err == nil {
			return fmt.Errorf("Permission on connection %s still exists.", connectionHandle)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccTenantConnectionPermission_Basic testAccCheckTenantConnectionPermissionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckTenantConnectionPermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		connectionHandle := rs.Primary.Attributes["connection_handle"]
		permissionId := rs.Primary.Attributes["permission_id"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		_, r, err = client.APIClient.TenantConnections.GetPermission(context.Background(), connectionHandle, permissionId).Execute()
		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Permission on connection %s in tenant not found.\nstatus: %d \nerr: %v", connectionHandle, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnectionPermission_Basic testAccCheckTenantConnectionPermissionExists %v", err)
			return err
		}
		return nil
	}
}

func testAccCheckTenantConnectionAccess(n, orgAvailable, orgNotAvailable string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		connectionHandle := rs.Primary.Attributes["handle"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		// The connection should be accessible by the organization with handle `orgAvailable`
		_, r, err = client.APIClient.OrgConnections.Get(context.Background(), orgAvailable, connectionHandle).Execute()
		// If there's an error and its a not found error, it means the connection is not available to the organization, fail the test
		if err != nil {
			if r.StatusCode == 404 {
				return fmt.Errorf("Connection %s in organization %s not found.\nstatus: %d \nerr: %v", connectionHandle, orgAvailable, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnectionPermission_Basic testAccCheckTenantConnectionAccess %v", err)
			return err
		}

		// The connection should not be accessible by the organization with handle `orgNotAvailable`
		_, r, err = client.APIClient.OrgConnections.Get(context.Background(), orgNotAvailable, connectionHandle).Execute()
		// If there's no error here, it means the connection has bern returned which is not expected, fail the test
		if err == nil {
			return fmt.Errorf("Connection %s should not be available to organization %s.\nstatus: %d \nerr: %v", connectionHandle, orgNotAvailable, r.StatusCode, r.Body)
		}

		return nil
	}
}
