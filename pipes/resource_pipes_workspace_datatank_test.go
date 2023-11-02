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
func TestAccUserWorkspaceDatatank_Basic(t *testing.T) {
	datatankResourceName := "pipes_workspace_datatank.test_datatank_fast_net"
	workspaceHandle := "workspace" + randomString(3)
	datatankHandle := "fast_net"
	datatankDescription := "Fast access to net data."
	updatedDatatankDescription := "Updated fast access to net data."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDatatankDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceDatatankConfig(workspaceHandle, datatankHandle, datatankDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceDatatankExists(workspaceHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "handle", datatankHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "description", datatankDescription),
				),
			},
			{
				ResourceName:            datatankResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceDatatankUpdateConfig(workspaceHandle, datatankHandle, updatedDatatankDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceDatatankExists(workspaceHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "handle", datatankHandle),
					resource.TestCheckResourceAttr(datatankResourceName, "description", updatedDatatankDescription),
				),
			},
			{
				ResourceName:            datatankResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccUserWorkspaceDatatankConfig(workspaceHandle, datatankHandle, description string) string {
	return fmt.Sprintf(`
	provider "pipes" {}

	resource "pipes_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "pipes_connection" "test_connection_net_1" {
		handle     = "net_1"
		plugin     = "net"
	}

	resource "pipes_connection" "test_connection_net_2" {
		handle     = "net_2"
		plugin     = "net"
	}
	
	resource "pipes_workspace_connection" "test_connection_net_1_association" {
	  workspace_handle  = pipes_workspace.test_workspace.handle
	  connection_handle = pipes_connection.test_connection_net_1.handle
	}

	resource "pipes_workspace_connection" "test_connection_net_2_association" {
		workspace_handle  = pipes_workspace.test_workspace.handle
		connection_handle = pipes_connection.test_connection_net_2.handle
	}

	resource "pipes_workspace_aggregator" "test_aggregator_all_net" {
		workspace = pipes_workspace.test_workspace.handle
		handle             = "all_net"
		plugin             = "net"
		connections        = ["*"]
	}
	
	resource "pipes_workspace_datatank" "test_datatank_fast_net" {
		workspace = pipes_workspace.test_workspace.handle
		handle 		  = "%s"
		description   = "%s"
	}`, workspaceHandle, datatankHandle, description)
}

func testAccUserWorkspaceDatatankUpdateConfig(workspaceHandle, datatankHandle, description string) string {
	return fmt.Sprintf(`
	provider "pipes" {}

	resource "pipes_workspace" "test_workspace" {
		handle = "%s"
	}

	resource "pipes_connection" "test_connection_net_1" {
		handle     = "net_1"
		plugin     = "net"
	}

	resource "pipes_connection" "test_connection_net_2" {
		handle     = "net_2"
		plugin     = "net"
	}
	
	resource "pipes_workspace_connection" "test_connection_net_1_association" {
	  workspace_handle  = pipes_workspace.test_workspace.handle
	  connection_handle = pipes_connection.test_connection_net_1.handle
	}

	resource "pipes_workspace_connection" "test_connection_net_2_association" {
		workspace_handle  = pipes_workspace.test_workspace.handle
		connection_handle = pipes_connection.test_connection_net_2.handle
	}

	resource "pipes_workspace_aggregator" "test_aggregator_all_net" {
		workspace = pipes_workspace.test_workspace.handle
		handle             = "all_net"
		plugin             = "net"
		connections        = ["*"]
	}
	
	resource "pipes_workspace_datatank" "test_datatank_fast_net" {
		workspace = pipes_workspace.test_workspace.handle
		handle 		  = "%s"
		description   = "%s"
	}`, workspaceHandle, datatankHandle, description)
}

func testAccCheckWorkspaceDatatankExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_datatank" {
				continue
			}

			datatankId := rs.Primary.Attributes["datatank_id"]
			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""

			var err error
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %s", err)
				}
				_, _, err = client.APIClient.UserWorkspaceDatatanks.Get(ctx, userHandle, workspaceHandle, datatankId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching datatank %s in user workspace with handle %s. %s", datatankId, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceDatatanks.Get(ctx, org, workspaceHandle, datatankId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching datatank %s in org workspace with handle %s. %s", datatankId, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceDatatankDestroy verifies the datatank has been deleted from the workspace
func testAccCheckWorkspaceDatatankDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_datatank" {
			continue
		}

		// Retrieve workspace handle and datatank id by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		datatankId := rs.Primary.Attributes["datatank_id"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceDatatanks.Get(ctx, userHandle, workspaceHandle, datatankId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceDatatanks.Get(ctx, org, workspaceHandle, datatankId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace Datatank %s/%s still exists", workspaceHandle, datatankId)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspaceDatatankDestroy testAccCheckUserWorkspaceDatatankDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspaceDatatankDestroy testAccCheckUserWorkspaceDatatankDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
