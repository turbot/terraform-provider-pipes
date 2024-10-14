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

// Test Suites
func TestAccUserWorkspaceFlowpipeMod_Basic(t *testing.T) {
	resourceName := "pipes_workspace_flowpipe_mod.aws_thrifty"
	workspaceHandle := "ws" + randomString(3)
	modPath := "github.com/turbot/flowpipe-mod-aws-thrifty"
	modAlias := "aws_thrifty"
	constraint := "*"
	newConstraint := ">v0.2.0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceFlowpipeModDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceFlowpipeModConfig(workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", constraint),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccUserWorkspaceFlowpipeModUpdateConfig(workspaceHandle, modPath, newConstraint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", newConstraint),
				),
			},
		},
	})
}

func TestAccOrgWorkspaceFlowpipeMod_Basic(t *testing.T) {
	resourceName := "pipes_workspace_flowpipe_mod.aws_thrifty"
	orgHandle := "terraformtest"
	workspaceHandle := "ws" + randomString(3)
	modPath := "github.com/turbot/flowpipe-mod-aws-thrifty"
	modAlias := "aws_thrifty"
	constraint := "*"
	newConstraint := ">v0.2.0"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceFlowpipeModDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgWorkspaceFlowpipeModConfig(orgHandle, workspaceHandle, modPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "organization", orgHandle),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", constraint),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
			{
				Config: testAccOrgWorkspaceFlowpipeModUpdateConfig(orgHandle, workspaceHandle, modPath, newConstraint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceFlowpipeModExists(workspaceHandle, modAlias),
					resource.TestCheckResourceAttr(resourceName, "organization", orgHandle),
					resource.TestCheckResourceAttr(resourceName, "alias", modAlias),
					resource.TestCheckResourceAttr(resourceName, "constraint", newConstraint),
				),
			},
		},
	})
}

func testAccUserWorkspaceFlowpipeModConfig(wsHandle, modPath string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "test_workspace" {
	handle = "%s"
}

resource "pipes_workspace_flowpipe_mod" "aws_thrifty" {
	workspace_handle = pipes_workspace.test_workspace.handle
	path = "%s"
}`, wsHandle, modPath)
}

func testAccUserWorkspaceFlowpipeModUpdateConfig(wsHandle, modPath, newConstraint string) string {
	return fmt.Sprintf(`
resource "pipes_workspace" "test_workspace" {
	handle = "%s"
}

resource "pipes_workspace_flowpipe_mod" "aws_thrifty" {
	workspace_handle = pipes_workspace.test_workspace.handle
	path = "%s"
	constraint = "%s"
}`, wsHandle, modPath, newConstraint)
}

func testAccOrgWorkspaceFlowpipeModConfig(orgHandle, wsHandle, modPath string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test"
}

resource "pipes_workspace" "test_workspace" {
	organization = pipes_organization.test_org.handle
	handle = "%s"
}

resource "pipes_workspace_flowpipe_mod" "aws_thrifty" {
	organization = pipes_organization.test_org.handle
	workspace_handle = pipes_workspace.test_workspace.handle
	path = "%s"
}`, orgHandle, wsHandle, modPath)
}

func testAccOrgWorkspaceFlowpipeModUpdateConfig(orgHandle, wsHandle, modPath, newConstraint string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org" {
	handle       = "%s"
	display_name = "Terraform Test"
}

resource "pipes_workspace" "test_workspace" {
	organization = pipes_organization.test_org.handle
	handle = "%s"
}

resource "pipes_workspace_flowpipe_mod" "aws_thrifty" {
	organization = pipes_organization.test_org.handle
	workspace_handle = pipes_workspace.test_workspace.handle
	path = "%s"
	constraint = "%s"
}`, orgHandle, wsHandle, modPath, newConstraint)
}

// testAccCheckWorkspaceFlowpipeModExists verifies the flowpipe mod resource exists
func testAccCheckWorkspaceFlowpipeModExists(workspaceHandle, modAlias string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_mod" {
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
				_, _, err = client.APIClient.UserWorkspaceFlowpipeMods.Get(ctx, actorHandle, workspaceHandle, modAlias).Execute()
				if err != nil {
					return fmt.Errorf("error fetching flowpipe mod %s in user workspace with handle %s. %s", modAlias, workspaceHandle, err)
				}
			} else {
				_, _, err = client.APIClient.OrgWorkspaceFlowpipeMods.Get(ctx, org, workspaceHandle, modAlias).Execute()
				if err != nil {
					return fmt.Errorf("error fetching flowpipe mod %s in org workspace with handle %s. %s", modAlias, workspaceHandle, err)
				}
			}
		}
		return nil
	}
}

// testAccCheckWorkspaceFlowpipeModDestroy verifies the flowpipe mod resource is destroyed
func testAccCheckWorkspaceFlowpipeModDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response
	var expectedStatusCode int

	client := testAccProvider.Meta().(*PipesClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "pipes_workspace_flowpipe_mod" {
			continue
		}

		workspaceHandle := res.Primary.Attributes["workspace_handle"]
		modAlias := res.Primary.Attributes["alias"]
		orgHandle := res.Primary.Attributes["organization"]
		isUser := orgHandle == ""

		if isUser {
			// user context returns 404 not found when it cannot access the mod
			expectedStatusCode = http.StatusNotFound
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %s", err)
			}
			_, r, err = client.APIClient.UserWorkspaceFlowpipeMods.Get(ctx, userHandle, workspaceHandle, modAlias).Execute()
		} else {
			// org context returns 403 forbidden when it cannot access the mod
			expectedStatusCode = http.StatusForbidden
			_, r, err = client.APIClient.OrgWorkspaceFlowpipeMods.Get(ctx, orgHandle, workspaceHandle, modAlias).Execute()
		}

		if err == nil {
			return fmt.Errorf("flowpipe mod %s:%s still exists", workspaceHandle, modAlias)
		}
		if r.StatusCode != expectedStatusCode {
			log.Printf("[DEBUG] testAccCheckWorkspaceFlowpipeModDestroy unexpected status code %d. Expected %d", r.StatusCode, expectedStatusCode)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}
	}
	return nil
}
