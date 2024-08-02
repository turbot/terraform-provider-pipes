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

func TestAccTenantConnectionFolder_Basic(t *testing.T) {
	resourceName := "pipes_tenant_connection_folder.folder1"
	title := "My Test Connection Folder"
	updatedTitle := "My Updated Test Connection Folder"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantConnectionFolderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConnectionFolderConfig(title),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTenantConnectionFolderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "title", title),
					resource.TestCheckResourceAttr(resourceName, "parent_id", ""),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccTenantConnectionFolderTitleUpdateConfig(updatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "title", updatedTitle),
					resource.TestCheckResourceAttr(resourceName, "parent_id", ""),
				),
			},
		},
	})
}

// configs
func testAccTenantConnectionFolderConfig(title string) string {
	return fmt.Sprintf(`
resource "pipes_tenant_connection_folder" "folder1" {
	title  = "%s"
}`, title)
}

func testAccTenantConnectionFolderTitleUpdateConfig(title string) string {
	return fmt.Sprintf(`
resource "pipes_tenant_connection_folder" "folder1" {
	title  = "%s"
}`, title)
}

// testAccCheckTenantConnectionFolderDestroy verifies the connection has been destroyed
func testAccCheckTenantConnectionFolderDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_tenant_connection_folder" {
			continue
		}

		// Retrieve connection by referencing it's state handle for API lookup
		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]

		_, r, err = client.APIClient.TenantConnectionFolders.Get(ctx, connectionFolderId).Execute()
		if err == nil {
			return fmt.Errorf("Connection Folder %s still exists in tenant.", connectionFolderId)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccTenantConnectionFolder_Basic testAccCheckTenantConnectionFolderDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckTenantConnectionFolderExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		_, r, err = client.APIClient.TenantConnectionFolders.Get(context.Background(), connectionFolderId).Execute()
		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection folder %s in tenant not found.\nstatus: %d \nerr: %v", connectionFolderId, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnectionFolder_Basic testAccCheckTenantConnectionFolderExists %v", err)
			return err
		}
		return nil
	}
}
