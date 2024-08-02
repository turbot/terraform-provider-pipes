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

func TestAccWorkspaceSchema_Basic(t *testing.T) {
	orgResource := "pipes_organization.test_org"
	workspaceResource1 := "pipes_workspace.test_workspace_1"
	workspaceResource2 := "pipes_workspace.test_workspace_2"
	connResource := "pipes_organization_connection.connection_1"
	permissionResource := "pipes_organization_connection_permission.permission_1"
	schemaResource := "pipes_workspace_schema.schema_1"
	orgHandle := "org" + randomString(5)
	workspaceHandle1 := "workspace" + randomString(6)
	workspaceHandle2 := "workspace" + randomString(6)
	connHandle := "aws" + randomString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceSchemaConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource),
					testAccCheckOrgWorkspaceExists(workspaceResource1),
					testAccCheckOrgWorkspaceExists(workspaceResource2),
					testAccCheckOrgConnectionExists(connResource),
					testAccCheckOrgConnectionPermissionExists(permissionResource),
					testAccCheckWorkspaceSchemaExists(schemaResource),
					resource.TestCheckResourceAttr(schemaResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(schemaResource, "organization", orgHandle),
					resource.TestCheckResourceAttr(schemaResource, "workspace", workspaceHandle1),
				),
			},
			{
				ResourceName: permissionResource,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceSchemaUpdateConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource),
					testAccCheckOrgWorkspaceExists(workspaceResource1),
					testAccCheckOrgWorkspaceExists(workspaceResource2),
					testAccCheckOrgConnectionExists(connResource),
					testAccCheckOrgConnectionPermissionExists(permissionResource),
					testAccCheckWorkspaceSchemaExists(schemaResource),
					resource.TestCheckResourceAttr(schemaResource, "connection_handle", connHandle),
					resource.TestCheckResourceAttr(schemaResource, "organization", orgHandle),
					resource.TestCheckResourceAttr(schemaResource, "workspace", workspaceHandle2),
				),
			},
		},
	})
}

// configs
func testAccWorkspaceSchemaConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle string) string {
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
}
	
resource "pipes_workspace_schema" "schema_1" {
	organization      = pipes_organization_connection_permission.permission_1.identity_handle
	workspace 	      = pipes_workspace.test_workspace_1.handle
	connection_handle = pipes_organization_connection.connection_1.handle
}`, orgHandle, workspaceHandle1, workspaceHandle2, connHandle)
}

func testAccWorkspaceSchemaUpdateConfig(orgHandle, workspaceHandle1, workspaceHandle2, connHandle string) string {
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
}
	
resource "pipes_workspace_schema" "schema_1" {
	organization      = pipes_organization_connection_permission.permission_1.identity_handle
	workspace 	      = pipes_workspace.test_workspace_2.handle
	connection_handle = pipes_organization_connection.connection_1.handle
}`, orgHandle, workspaceHandle1, workspaceHandle2, connHandle)
}

// testAccCheckWorkspaceSchemaDestroy verifies the connection permission has been destroyed
func testAccCheckWorkspaceSchemaDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_schema" {
			continue
		}

		orgHandle := rs.Primary.Attributes["organization"]
		workspaceHandle := rs.Primary.Attributes["workspace"]
		schemaName := rs.Primary.Attributes["connection_handle"]

		_, r, err = client.APIClient.OrgWorkspaceSchemas.Get(ctx, orgHandle, workspaceHandle, schemaName).Execute()
		if err == nil {
			return fmt.Errorf("Schema %s is still attached to workspace %s of org %s.", schemaName, workspaceHandle, orgHandle)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccWorkspaceSchema_Basic testAccCheckWorkspaceSchemaDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckWorkspaceSchemaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		orgHandle := rs.Primary.Attributes["organization"]
		workspaceHandle := rs.Primary.Attributes["workspace"]
		schemaName := rs.Primary.Attributes["connection_handle"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		_, r, err = client.APIClient.OrgWorkspaceSchemas.Get(context.Background(), orgHandle, workspaceHandle, schemaName).Execute()
		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Schema %s is not attached to workspace %s of org %s.\nstatus: %d \nerr: %v", schemaName, workspaceHandle, orgHandle, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccWorkspaceSchema_Basic testAccCheckWorkspaceSchemaExists %v", err)
			return err
		}
		return nil
	}
}
