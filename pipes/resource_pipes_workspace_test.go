package pipes

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// test suites
func TestAccUserWorkspace_Basic(t *testing.T) {
	resourceName := "pipes_workspace.test"
	workspaceHandle := "workspace" + randomString(3)
	newWorkspaceHandle := "workspace" + randomString(4)
	workspaceInstanceType := "db1.small"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceConfig(workspaceHandle, workspaceInstanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserWorkspaceExists("pipes_workspace.test"),
					resource.TestCheckResourceAttr("pipes_workspace.test", "handle", workspaceHandle),
					resource.TestCheckResourceAttr("pipes_workspace.test", "instance_type", workspaceInstanceType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceUpdateHandleConfig(newWorkspaceHandle, workspaceInstanceType),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pipes_workspace.test", "handle", newWorkspaceHandle),
					resource.TestCheckResourceAttr("pipes_workspace.test", "instance_type", workspaceInstanceType),
				),
			},
		},
	})
}

// configs
func testAccUserWorkspaceConfig(workspaceHandle, workspaceInstanceType string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "test" {
	handle = "%s"
	instance_type = "%s"
}`, workspaceHandle, workspaceInstanceType)
}

func testAccUserWorkspaceUpdateHandleConfig(newWorkspaceHandle, workspaceInstanceType string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "test" {
	handle = "%s"
	instance_type = "%s"
}`, newWorkspaceHandle, workspaceInstanceType)
}

// helper functions
func testAccCheckUserWorkspaceExists(resource string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		client := testAccProvider.Meta().(*PipesClient)

		// Get user handle
		userData, _, userErr := client.APIClient.Actors.Get(ctx).Execute()
		if userErr != nil {
			return fmt.Errorf("error fetching user handle. %s", userErr)
		}

		_, _, err := client.APIClient.UserWorkspaces.Get(ctx, userData.Handle, rs.Primary.ID).Execute()
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckUserWorkspaceDestroy(s *terraform.State) error {
	ctx := context.Background()
	client := testAccProvider.Meta().(*PipesClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "pipes_workspace" {
			// Get user handle
			userData, _, userErr := client.APIClient.Actors.Get(ctx).Execute()
			if userErr != nil {
				return fmt.Errorf("error fetching user handle. %s", userErr)
			}

			_, r, err := client.APIClient.UserWorkspaces.Get(ctx, userData.Handle, rs.Primary.ID).Execute()
			if err == nil {
				return fmt.Errorf("Workspace still exists")
			}

			if r.StatusCode != 404 {
				return fmt.Errorf("expected 'no content' error, got %s", err)
			}
		}
	}

	return nil
}
