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

func TestAccWorkspaceConnection_Basic(t *testing.T) {
	resourceName := "pipes_workspace_connection.test_conn"
	workspaceHandle := "workspace" + randomString(6)
	connHandle := "aws_" + randomString(4)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConnectionConfig(workspaceHandle, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
		},
	})
}

func TestAccOrgWorkspaceConnection_Basic(t *testing.T) {
	resourceName := "pipes_workspace_connection.test_org"
	orgName := "terraform-" + randomString(11)
	workspaceHandle := "workspace" + randomString(5)
	connHandle := "aws_" + randomString(3)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceConnectionConfig(orgName, workspaceHandle, connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgName),
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckWorkspaceConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "workspace", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
				),
			},
		},
	})
}

// User Workspace Connection association config
func testAccWorkspaceConnectionConfig(workspace string, conn string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "test_conn" {
  handle = "%s"
}

resource "pipes_workspace_connection" "test_conn" {
	workspace  = pipes_workspace.test_conn.handle
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}
`, workspace, conn)
}

// Organization Workspace Connection association config
func testAccOrgWorkspaceConnectionConfig(org string, workspace string, conn string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_workspace" "test_workspace" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_workspace_connection" "test_workspace_connection" {
	organization = pipes_organization.test_org.handle
	workspace    = pipes_workspace.test_workspace.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}
`, org, workspace, conn)
}

// testAccCheckWorkspaceConnectionDestroy verifies the workspace connection association has been destroyed
func testAccCheckWorkspaceConnectionDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_connection" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["handle"]
		workspaceHandle := rs.Primary.Attributes["workspace"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceConnections.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceConnections.Get(ctx, org, workspaceHandle, connectionHandle).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Connection association %s:%s still exists", workspaceHandle, connectionHandle)
		}

		// If the error is equivalent to 404 not found, the workspace connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccWorkspaceConnection_Basic testAccCheckWorkspaceConnectionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckWorkspaceConnectionExists(n string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		var err error

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connectionHandle := rs.Primary.Attributes["handle"]
		workspaceHandle := rs.Primary.Attributes["workspace"]

		client := testAccProvider.Meta().(*PipesClient)

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaceConnections.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading user workspace connection: %s:%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaceConnections.Get(ctx, org, workspaceHandle, connectionHandle).Execute()
			if err != nil {
				return fmt.Errorf("error reading organization workspace connection: %s:%s.\nerr: %s", workspaceHandle, connectionHandle, err)
			}
		}

		return nil
	}
}

func testAccCheckTestWorkspaceExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace" {
				continue
			}

			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var actorHandle string
				actorHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspaces.Get(ctx, actorHandle, workspaceHandle).Execute()
				if err != nil {
					return fmt.Errorf("error fetching user workspace with handle %s. %s", workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaces.Get(ctx, org, workspaceHandle).Execute()
				if err != nil {
					return fmt.Errorf("error fetching org workspace with handle %s. %s", workspaceHandle, err)
				}
			}
		}
		return nil
	}
}
