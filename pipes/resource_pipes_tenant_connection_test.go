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

func TestAccTenantConnection_Basic(t *testing.T) {
	resourceName := "pipes_tenant_connection.test"
	connHandle := "aws_" + randomString(5)
	newHandle := "aws_" + randomString(6)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConnectionConfig(connHandle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "handle", connHandle),
					resource.TestCheckResourceAttr(resourceName, "plugin", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
					testCheckJSONString(resourceName, "config", `{"access_key":"redacted","regions":["us-east-1"],"secret_key":"redacted"}`),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccTenantConnectionHandleUpdateConfig(newHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pipes_connection.test", "handle", newHandle),
					resource.TestCheckResourceAttr(resourceName, "config", "{\n \"access_key\": \"redacted\",\n \"regions\": [\n  \"us-east-2\",\n  \"us-east-1\"\n ],\n \"secret_key\": \"redacted\"\n}"),
					testCheckJSONString(resourceName, "config", `{"access_key":"redacted","regions":[""us-east-2","us-east-1"],"secret_key":"redacted"}`),
				),
			},
		},
	})
}

// configs
func testAccTenantConnectionConfig(connHandle string) string {
	return fmt.Sprintf(`
provider "pipes" {}

resource "pipes_tenant_connection" "test" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}`, connHandle)
}

func testAccTenantConnectionHandleUpdateConfig(newHandle string) string {
	return fmt.Sprintf(`
provider "pipes" {}

resource "pipes_tenant_connection" "test" {
	handle     = "%s"
	plugin     = "aws"
	config = jsonencode({
		regions    = ["us-east-2", "us-east-1"]
		access_key = "redacted"
		secret_key = "redacted"
	})
}`, newHandle)
}

// testAccCheckTenantConnectionDestroy verifies the connection has been destroyed
func testAccCheckTenantConnectionDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_tenant_connection" {
			continue
		}

		// Retrieve connection by referencing it's state handle for API lookup
		connectionHandle := rs.Primary.Attributes["handle"]

		_, r, err = client.APIClient.TenantConnections.Get(ctx, connectionHandle).Execute()
		if err == nil {
			return fmt.Errorf("Connection %s still exists in tenant.", connectionHandle)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccTenantConnection_Basic testAccCheckTenantConnectionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckTenantConnectionExists(n string) resource.TestCheckFunc {
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

		var r *http.Response
		var err error

		_, r, err = client.APIClient.TenantConnections.Get(context.Background(), connectionHandle).Execute()
		if err != nil {
			return fmt.Errorf("testAccCheckTenantConnectionExists.\n Get tenant connection error: %v", decodeResponse(r))
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection %s in tenant not found.\nstatus: %d \nerr: %v", connectionHandle, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnection_Basic testAccCheckTenantConnectionExists %v", err)
			return err
		}
		return nil
	}
}
