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

func TestAccWorkspaceConnectionFolder_Basic(t *testing.T) {
	folderResourceName := "pipes_workspace_connection_folder.folder1"
	workspaceHandle := "workspace" + randomString(6)
	title := "My Workspace test connection folder"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceConnectionFolderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConnectionFolderConfig(workspaceHandle, title),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckWorkspaceConnectionFolderExists(folderResourceName),
					resource.TestCheckResourceAttr(folderResourceName, "workspace", workspaceHandle),
					resource.TestCheckResourceAttr(folderResourceName, "title", title),
					resource.TestCheckResourceAttr(folderResourceName, "parent_id", ""),
				),
			},
			{
				ResourceName: folderResourceName,
				ImportState:  true,
			},
		},
	})
}

func TestAccOrgWorkspaceConnectionFolder_Basic(t *testing.T) {
	folderResourceName := "pipes_workspace_connection_folder.folder1"
	orgResourceName := "pipes_organization.test_org"
	orgHandle := "terraform-" + randomString(11)
	workspaceHandle := "workspace" + randomString(5)
	title := "My Org Workspace test connection folder"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceConnectionFolderConfig(orgHandle, workspaceHandle, title),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResourceName),
					testAccCheckTestWorkspaceExists(workspaceHandle),
					testAccCheckWorkspaceConnectionFolderExists(folderResourceName),
					resource.TestCheckResourceAttr(folderResourceName, "workspace", workspaceHandle),
					resource.TestCheckResourceAttr(folderResourceName, "title", title),
					resource.TestCheckResourceAttr(folderResourceName, "parent_id", ""),
				),
			},
		},
	})
}

// User Workspace Connection association config
func testAccWorkspaceConnectionFolderConfig(workspace, title string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "workspace1" {
  handle = "%s"
}

resource "pipes_workspace_connection_folder" "folder1" {
	workspace  = pipes_workspace.workspace1.handle
	title = "%s"
}
`, workspace, title)
}

// Organization Workspace Connection association config
func testAccOrgWorkspaceConnectionFolderConfig(org, workspace, title string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_workspace" "test_workspace" {
	organization = pipes_organization.test_org.handle
	handle       = "%s"
}

resource "pipes_workspace_connection_folder" "folder1" {
	organization = pipes_organization.test_org.handle
	workspace    = pipes_workspace.test_workspace.handle
	title        = "%s"
}
`, org, workspace, title)
}

// testAccCheckWorkspaceConnectionFolderDestroy verifies the workspace connection association has been destroyed
func testAccCheckWorkspaceConnectionFolderDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_connection_folder" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]
		workspaceHandle := rs.Primary.Attributes["workspace"]

		// Retrieve organization
		orgHandle := rs.Primary.Attributes["organization"]
		isUser := orgHandle == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceConnectionFolders.Get(ctx, actorHandle, workspaceHandle, connectionFolderId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceConnectionFolders.Get(ctx, orgHandle, workspaceHandle, connectionFolderId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace connection folder %s:%s still exists", workspaceHandle, connectionFolderId)
		}

		// If the error is equivalent to 404 not found, the workspace connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccWorkspaceConnectionFolder_Basic testAccCheckWorkspaceConnectionFolderDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckWorkspaceConnectionFolderExists(n string) resource.TestCheckFunc {
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
		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]
		workspaceHandle := rs.Primary.Attributes["workspace"]

		client := testAccProvider.Meta().(*PipesClient)

		// Retrieve organization
		orgHandle := rs.Primary.Attributes["organization"]
		isUser := orgHandle == ""

		if isUser {
			var actorHandle string
			actorHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, _, err = client.APIClient.UserWorkspaceConnectionFolders.Get(ctx, actorHandle, workspaceHandle, connectionFolderId).Execute()
			if err != nil {
				return fmt.Errorf("error reading user workspace connection folder: %s:%s.\nerr: %s", workspaceHandle, connectionFolderId, err)
			}
		} else {
			_, _, err = client.APIClient.OrgWorkspaceConnectionFolders.Get(ctx, orgHandle, workspaceHandle, connectionFolderId).Execute()
			if err != nil {
				return fmt.Errorf("error reading organization workspace connection: %s:%s.\nerr: %s", workspaceHandle, connectionFolderId, err)
			}
		}

		return nil
	}
}
