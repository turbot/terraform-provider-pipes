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
func TestAccUserWorkspaceDatatankTable_Basic(t *testing.T) {
	datatankTableResourceName := "pipes_workspace_datatank_table.test_datatank_table_net_certificate"
	workspaceHandle := "workspace" + randomString(3)
	datatankHandle := "fast_net"
	name := "net_certificate"
	tableType := "table"
	partPer := "connection"
	sourceSchema := "all_net"
	sourceTable := "net_certificate"
	frequency := `
		{
			"type": "interval",
			"schedule": "daily"
		}
	`
	updatedFrequency := `
		{
			"type": "interval",
			"schedule": "hourly"
		}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceDatatankTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceDatatankTableConfig(workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, frequency),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceDatatankTableExists(workspaceHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "datatank_handle", datatankHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "name", name),
					resource.TestCheckResourceAttr(datatankTableResourceName, "type", tableType),
					resource.TestCheckResourceAttr(datatankTableResourceName, "part_per", partPer),
					resource.TestCheckResourceAttr(datatankTableResourceName, "source_schema", sourceSchema),
					resource.TestCheckResourceAttr(datatankTableResourceName, "source_table", sourceTable),
					TestJSONFieldEqual(t, datatankTableResourceName, "frequency", frequency),
				),
			},
			{
				ResourceName:            datatankTableResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "frequency", "freshness", "migrating_freshness"},
			},
			{
				Config: testAccUserWorkspaceDatatankTableUpdateConfig(workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, updatedFrequency),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceDatatankTableExists(workspaceHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "datatank_handle", datatankHandle),
					resource.TestCheckResourceAttr(datatankTableResourceName, "name", name),
					resource.TestCheckResourceAttr(datatankTableResourceName, "type", tableType),
					resource.TestCheckResourceAttr(datatankTableResourceName, "part_per", partPer),
					resource.TestCheckResourceAttr(datatankTableResourceName, "source_schema", sourceSchema),
					resource.TestCheckResourceAttr(datatankTableResourceName, "source_table", sourceTable),
					TestJSONFieldEqual(t, datatankTableResourceName, "frequency", updatedFrequency),
				),
			},
			{
				ResourceName:            datatankTableResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at", "frequency", "freshness", "migrating_freshness"},
			},
		},
	})
}

func testAccUserWorkspaceDatatankTableConfig(workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, frequency string) string {
	return fmt.Sprintf(`
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
		workspace_handle = pipes_workspace.test_workspace.handle
		handle 		     = "%s"
	}
	
	resource "pipes_workspace_datatank_table" "test_datatank_table_net_certificate" {
		workspace_handle = pipes_workspace.test_workspace.handle
		datatank_handle  = pipes_workspace_datatank.test_datatank_fast_net.handle
		name 		     = "%s"
		type             = "%s"
		part_per         = "%s"
		source_schema	 = "%s"
		source_table	 = "%s"
		frequency        = jsonencode(%s)

		depends_on = [pipes_workspace_aggregator.test_aggregator_all_net]
	}`, workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, frequency)
}

func testAccUserWorkspaceDatatankTableUpdateConfig(workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, frequency string) string {
	return fmt.Sprintf(`
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
		workspace_handle = pipes_workspace.test_workspace.handle
		handle 		     = "%s"
	}
	
	resource "pipes_workspace_datatank_table" "test_datatank_table_net_certificate" {
		workspace_handle = pipes_workspace.test_workspace.handle
		datatank_handle  = pipes_workspace_datatank.test_datatank_fast_net.handle
		name 		     = "%s"
		type             = "%s"
		part_per         = "%s"
		source_schema	 = "%s"
		source_table	 = "%s"
		frequency        = jsonencode(%s)

		depends_on = [pipes_workspace_aggregator.test_aggregator_all_net]
	}`, workspaceHandle, datatankHandle, name, tableType, partPer, sourceSchema, sourceTable, frequency)
}

func testAccCheckWorkspaceDatatankTableExists(workspaceHandle string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_datatank" {
				continue
			}

			datatankId := rs.Primary.Attributes["datatank_id"]
			datatankTableId := rs.Primary.Attributes["datatank_table_id"]
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
				_, _, err = client.APIClient.UserWorkspaceDatatankTables.Get(ctx, userHandle, workspaceHandle, datatankId, datatankTableId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching datatank %s in user workspace with handle %s. %s", datatankId, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceDatatankTables.Get(ctx, org, workspaceHandle, datatankId, datatankTableId).Execute()
				if err != nil {
					return fmt.Errorf("error fetching datatank %s in org workspace with handle %s. %s", datatankId, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceDatatankTableDestroy verifies the datatank has been deleted from the workspace
func testAccCheckWorkspaceDatatankTableDestroy(s *terraform.State) error {
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
		datatankTableId := rs.Primary.Attributes["datatank_table_id"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceDatatankTables.Get(ctx, userHandle, workspaceHandle, datatankId, datatankTableId).Execute()
		} else {
			_, r, err = client.APIClient.OrgWorkspaceDatatankTables.Get(ctx, org, workspaceHandle, datatankId, datatankTableId).Execute()
		}
		if err == nil {
			return fmt.Errorf("Workspace DatatankTable %s/%s still exists", workspaceHandle, datatankId)
		}

		if isUser {
			if r.StatusCode != 404 {
				log.Printf("[INFO] testAccCheckWorkspaceDatatankTableDestroy testAccCheckUserWorkspaceDatatankTableDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		} else {
			if r.StatusCode != 403 {
				log.Printf("[INFO] testAccCheckWorkspaceDatatankTableDestroy testAccCheckUserWorkspaceDatatankTableDestroy %v", err)
				return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
			}
		}

	}

	return nil
}
