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

func TestAccOrgConnectionPermission_Basic(t *testing.T) {
	orgResource := "pipes_organization.test_org"
	workspaceResource1 := "pipes_workspace.test_workspace_1"
	workspaceResource2 := "pipes_workspace.test_workspace_2"
	connResource := "pipes_organization_connection.connection_1"
	permissionResource := "pipes_organization_connection_permission.permission_1"
	orgHandle := "org" + randomString(5)
	workspaceHandle1 := "workspace" + randomString(6)
	workspaceHandle2 := "workspace" + randomString(6)
	connHandle := "aws" + randomString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgConnectionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionPermissionConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource),
					testAccCheckOrgWorkspaceExists(workspaceResource1),
					testAccCheckOrgWorkspaceExists(workspaceResource2),
					testAccCheckOrgConnectionExists(connResource),
					testAccCheckOrgConnectionPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle),
					resource.TestCheckResourceAttr(permissionResource, "workspace_handle", workspaceHandle1),
					testAccCheckOrgConnectionAccess(connResource, orgHandle, workspaceHandle1, workspaceHandle2),
				),
			},
			{
				ResourceName: permissionResource,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccOrgConnectionPermissionUpdateConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource),
					testAccCheckOrgWorkspaceExists(workspaceResource1),
					testAccCheckOrgWorkspaceExists(workspaceResource2),
					testAccCheckOrgConnectionExists(connResource),
					testAccCheckOrgConnectionPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle),
					resource.TestCheckResourceAttr(permissionResource, "workspace_handle", workspaceHandle2),
					testAccCheckOrgConnectionAccess(connResource, orgHandle, workspaceHandle2, workspaceHandle1),
				),
			},
		},
	})
}

// configs
func testAccOrgConnectionPermissionConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_workspace" "test_workspace_1" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_workspace" "test_workspace_2" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_organization_connection" "connection_1" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}

resource "pipes_organization_connection_permission" "permission_1" {
	organization      = pipes_organization.test_org.handle
	connection_handle = pipes_organization_connection.connection_1.handle
	identity_handle   = pipes_organization.test_org.handle
	workspace_handle  = pipes_workspace.test_workspace_1.handle
}`, orgHandle, workspaceHandle1, workspaceHandle2, connHandle)
}

func testAccOrgConnectionPermissionUpdateConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_workspace" "test_workspace_1" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_workspace" "test_workspace_2" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_organization_connection" "connection_1" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}

resource "pipes_organization_connection_permission" "permission_1" {
	organization      = pipes_organization.test_org.handle
	connection_handle = pipes_organization_connection.connection_1.handle
	identity_handle   = pipes_organization.test_org.handle
	workspace_handle  = pipes_workspace.test_workspace_2.handle
}`, orgHandle, workspaceHandle1, workspaceHandle2, connHandle)
}

// testAccCheckTenantConnectionPermissionDestroy verifies the connection permission has been destroyed
func testAccCheckOrgConnectionPermissionDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_organization_connection_permission" {
			continue
		}

		orgHandle := rs.Primary.Attributes["organization"]
		connectionHandle := rs.Primary.Attributes["connection_handle"]
		permissionId := rs.Primary.Attributes["permission_id"]

		_, r, err = client.APIClient.OrgConnections.GetPermission(ctx, orgHandle, connectionHandle, permissionId).Execute()
		if err == nil {
			return fmt.Errorf("Permission on connection %s of org %s still exists.", connectionHandle, orgHandle)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccOrgConnectionPermission_Basic testAccCheckOrgConnectionPermissionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckOrgConnectionPermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		orgHandle := rs.Primary.Attributes["organization"]
		connectionHandle := rs.Primary.Attributes["connection_handle"]
		permissionId := rs.Primary.Attributes["permission_id"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		_, r, err = client.APIClient.OrgConnections.GetPermission(context.Background(), orgHandle, connectionHandle, permissionId).Execute()
		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Permission on connection %s of org %s in tenant not found.\nstatus: %d \nerr: %v", connectionHandle, orgHandle, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccOrgConnectionPermission_Basic testAccCheckOrgConnectionPermissionExists %v", err)
			return err
		}
		return nil
	}
}

func testAccCheckOrgConnectionAccess(n, orgHandle, workspaceAvailable, workspaceNotAvailable string) resource.TestCheckFunc {
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
		_, r, err = client.APIClient.OrgWorkspaceConnections.Get(context.Background(), orgHandle, workspaceAvailable, connectionHandle).Execute()
		// If there's an error and its a not found error, it means the connection is not available to the organization, fail the test
		if err != nil {
			if r.StatusCode == 404 {
				return fmt.Errorf("Connection %s is not found in workspace %s of organization %s.\nstatus: %d \nerr: %v", connectionHandle, workspaceAvailable, orgHandle, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccOrgConnectionPermission_Basic testAccCheckOrgConnectionAccess %v", err)
			return err
		}

		// The connection should not be accessible by the organization with handle `orgNotAvailable`
		_, r, err = client.APIClient.OrgWorkspaceConnections.Get(context.Background(), orgHandle, workspaceNotAvailable, connectionHandle).Execute()
		// If there's no error here, it means the connection has bern returned which is not expected, fail the test
		if err == nil {
			return fmt.Errorf("Connection %s should not be available in workspace %s of organization %s.\nstatus: %d \nerr: %v", connectionHandle, workspaceNotAvailable, orgHandle, r.StatusCode, r.Body)
		}

		return nil
	}
}
