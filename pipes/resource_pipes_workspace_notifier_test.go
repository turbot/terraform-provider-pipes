package pipes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"net/http"
	"testing"
)

func TestAccUserWorkspaceNotifier_Basic(t *testing.T) {
	resourceName := "pipes_workspace_notifier.slack_general"
	workspaceHandle := "abc"
	integrationHandle := "slack"
	notifierName := "slack_general"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWorkspaceNotifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceNotifierConfig(workspaceHandle, integrationHandle, notifierName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceNotifierExists(workspaceHandle, notifierName),
					resource.TestCheckResourceAttr(resourceName, "name", notifierName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccUserWorkspaceNotifierConfig(workspaceHandle, integrationHandle, notifierName string) string {
	return fmt.Sprintf(`
data "pipes_integration" "slack" {
	handle = "%s"
}

resource "pipes_workspace_notifier" "slack_general" {
	workspace = "%s"
	name = "%s"
	notifies = jsonencode([{
		"type": "slack",
		"channel": "general",
		"integration": data.pipes_integration.slack.integration_id,
    }])
	state = "enabled"
}
`, integrationHandle, workspaceHandle, notifierName)
}

func testAccCheckWorkspaceNotifierExists(workspaceHandle, notifierName string) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*PipesClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pipes_workspace_notifier" {
				continue
			}

			var err error

			// Retrieve organization
			org := rs.Primary.Attributes["organization"]
			isUser := org == ""
			if isUser {
				var userHandle string
				userHandle, _, err = getUserHandler(ctx, client)
				if err != nil {
					return fmt.Errorf("error fetching user handle. %v", err)
				}
				_, _, err = client.APIClient.UserWorkspaceNotifiers.Get(ctx, userHandle, workspaceHandle, notifierName).Execute()
			} else {
				_, _, err = client.APIClient.OrgWorkspaceNotifiers.Get(ctx, org, workspaceHandle, notifierName).Execute()
			}

			if err != nil {
				return fmt.Errorf("error fetching workspace notifier. %v", err)
			}
		}

		return nil
	}
}

func testAccCheckWorkspaceNotifierDestroy(s *terraform.State) error {
	ctx := context.Background()
	var err error
	var r *http.Response
	var expectedStatusCode int

	client := testAccProvider.Meta().(*PipesClient)

	for _, res := range s.RootModule().Resources {
		if res.Type != "pipes_workspace_notifier" {
			continue
		}

		workspaceHandle := res.Primary.Attributes["workspace"]
		notifierName := res.Primary.Attributes["name"]
		orgHandle := res.Primary.Attributes["organization"]
		isUser := orgHandle == ""

		if isUser {
			expectedStatusCode = http.StatusNotFound
			var userHandle string
			userHandle, _, err = getUserHandler(ctx, client)
			if err != nil {
				return fmt.Errorf("error fetching user handle. %v", err)
			}
			_, r, err = client.APIClient.UserWorkspaceNotifiers.Get(ctx, userHandle, workspaceHandle, notifierName).Execute()
		} else {
			expectedStatusCode = http.StatusForbidden
			_, r, err = client.APIClient.OrgWorkspaceNotifiers.Get(ctx, orgHandle, workspaceHandle, notifierName).Execute()
		}

		if err == nil {
			return fmt.Errorf("workspace notifier %s:%s still exists. %v", workspaceHandle, notifierName, err)
		}

		if r.StatusCode != expectedStatusCode {
			log.Printf("\n[DEBUG] testAccCheckWorkspaceNotifierDestroy unexpected status code %d. Expected %d", r.StatusCode, expectedStatusCode)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}
	}

	return nil
}
