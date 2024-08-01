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

func TestAccOrgConnection_Basic(t *testing.T) {
	resourceName := "pipes_organization_connection.test_org"
	orgHandle := "terraform" + randomString(9)
	connHandle := "aws_" + randomString(7)
	newHandle := "aws_" + randomString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionConfig(connHandle, orgHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgHandle),
					testAccCheckOrgConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
				),
			},
			{
				Config: testAccOrgConnectionUpdateConfig(newHandle, orgHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pipes_organization_connection.test_org", "handle", newHandle),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-2\",\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
				),
			},
		},
	})
}

func testAccOrgConnectionConfig(connHandle string, orgHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_organization_connection" "test_org" {
	organization = pipes_organization.test.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions      = ["us-east-1"]
		access_key   = "redacted"
		secret_key   = "redacted"
	})
}`, orgHandle, connHandle)
}

func testAccOrgConnectionUpdateConfig(newHandle string, orgHandle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_organization_connection" "test_org" {
	organization = pipes_organization.test.handle
	handle       = "%s"
	plugin       = "aws"
	config = jsonencode({
		regions      = ["us-east-2", "us-east-1"]
		access_key   = "redacted"
		secret_key   = "redacted"
	})
}`, orgHandle, newHandle)
}

func testAccCheckOrgConnectionExists(n string) resource.TestCheckFunc {
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

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]

		var r *http.Response
		var err error
		_, r, err = client.APIClient.OrgConnections.Get(context.Background(), org, connectionHandle).Execute()
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection %s in organization %s not found.\nstatus: %d \nerr: %v", connectionHandle, org, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccOrgConnection_Basic testAccCheckConnectionExists %v", err)
			return err
		}

		return nil
	}
}
