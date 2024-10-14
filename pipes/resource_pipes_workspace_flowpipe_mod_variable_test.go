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

// Test case for user workspace only -
// Test case assumes that a user workspace already exists in the env of the handle abc
// TODO - Add the workspace creation and destruction logic as part of this test case.

func TestAccUserWorkspaceFlowpipeModVariable_Number(t *testing.T) {
	resourceName := "pipes_workspace_flowpipe_mod_variable.max_concurrency"
	modPath := "github.com/turbot/flowpipe-mod-aws-thrifty"
	modAlias := "aws_thrifty"
	variableName := "max_concurrency"
	defaultValue := "1"
	setting := "5"
	updatedSetting := "2"

	workspaceHandle := "abc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceFlowpipeModVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceFlowpipeModVariableConfig(workspaceHandle, modPath, variableName, setting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", setting),
					resource.TestCheckResourceAttr(resourceName, "value", setting),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceFlowpipeModVariableConfig(workspaceHandle, modPath, variableName, updatedSetting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", updatedSetting),
					resource.TestCheckResourceAttr(resourceName, "value", updatedSetting),
				),
			},
		},
	})
}

func TestAccUserWorkspaceFlowpipeModVariable_String(t *testing.T) {
	resourceName := "pipes_workspace_flowpipe_mod_variable.notification_level"
	modPath := "github.com/turbot/flowpipe-mod-aws-thrifty"
	modAlias := "aws_thrifty"
	variableName := "notification_level"
	defaultValue := "info"
	setting := "error"
	updatedSetting := "verbose"
	workspaceHandle := "abc"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceFlowpipeModVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceFlowpipeModVariableConfig(workspaceHandle, modPath, variableName, setting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", setting),
					resource.TestCheckResourceAttr(resourceName, "value", setting),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceFlowpipeModVariableConfig(workspaceHandle, modPath, variableName, updatedSetting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModVariableExists(workspaceHandle, modAlias, variableName),
					resource.TestCheckResourceAttr(resourceName, "workspace_handle", workspaceHandle),
					resource.TestCheckResourceAttr(resourceName, "mod_alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "name", variableName),
					resource.TestCheckResourceAttr(resourceName, "default_value", defaultValue),
					resource.TestCheckResourceAttr(resourceName, "setting_value", updatedSetting),
					resource.TestCheckResourceAttr(resourceName, "value", updatedSetting),
				),
			},
		},
	})
}

// testAccUserWorkspaceFlowpipeModVariableConfig returns the configuration for a user workspace flowpipe mod variable
func testAccUserWorkspaceFlowpipeModVariableConfig(workspaceHandle, modPath, variableName, setting string) string {
	return fmt.Sprintf(`
resource "pipes_workspace_flowpipe_mod" "aws_thrifty" {
	workspace_handle = "%s"
	path = "%s"
}

resource "pipes_workspace_flowpipe_mod_variable" "%s" {
	workspace_handle = "%s"
	mod_alias = pipes_workspace_flowpipe_mod.aws_thrifty.alias
	name = "%s"
	setting_value = %q
}`, workspaceHandle, modPath, variableName, workspaceHandle, variableName, setting)
}

// testAccCheckWorkspaceFlowpipeModVariableExists verifies the mod variable exists in the workspace
func testAccCheckWorkspaceFlowpipeModVariableExists(workspaceHandle, modAlias, variableName string) resource.TestCheckFunc {
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
				_, _, err = client.APIClient.UserWorkspaceFlowpipeModVariables.GetSetting(ctx, actorHandle, workspaceHandle, modAlias, variableName).Execute()
				if err != nil {
					return fmt.Errorf("error fetching variable %s in flowpipe mod %s for user workspace with handle %s. %s", variableName, modAlias, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.GetSetting(ctx, org, workspaceHandle, modAlias, variableName).Execute()
				if err != nil {
					return fmt.Errorf("error fetching variable %s in flowpipe mod %s for org workspace with handle %s. %s", variableName, modAlias, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceFlowpipeModVariableDestroy verifies the mod has been destroyed in the workspace
func testAccCheckWorkspaceFlowpipeModVariableDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response
	var expectedStatusCode int

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each managed resource is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_workspace_mod_variable" {
			continue
		}

		// Retrieve workspace and connection handle by referencing it's state handle for API lookup
		workspaceHandle := rs.Primary.Attributes["workspace_handle"]
		modAlias := rs.Primary.Attributes["mod_alias"]
		variableName := rs.Primary.Attributes["name"]

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]
		isUser := org == ""

		if isUser {
			// user context returns 404 not found when it cannot access the mod
			expectedStatusCode = http.StatusNotFound
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceFlowpipeModVariables.GetSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Execute()
		} else {
			// org context returns 403 forbidden when it cannot access the mod
			expectedStatusCode = http.StatusForbidden
			_, r, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.GetSetting(ctx, org, workspaceHandle, modAlias, variableName).Execute()
		}
		if err == nil {
			return fmt.Errorf("flowpipe mod variable %s:%s:%s still exists", workspaceHandle, modAlias, variableName)
		}

		if r.StatusCode != expectedStatusCode {
			log.Printf("[DEBUG] testAccCheckWorkspaceFlowpipeModVariableDestroy unexpected status code %d. Expected %d", r.StatusCode, expectedStatusCode)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}
	}

	return nil
}
