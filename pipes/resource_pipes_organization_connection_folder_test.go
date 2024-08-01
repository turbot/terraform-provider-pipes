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

func TestAccOrgConnectionFolder_Basic(t *testing.T) {
	folderName := "pipes_organization_connection_folder.folder1"
	orgName := "pipes_organization.test"
	orgHandle := "terraform" + randomString(9)
	title := "My Test Connection Folder"
	updatedTitle := "My Updated Test Connection Folder"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOrgConnectionFolderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgConnectionFolderConfig(orgHandle, title),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgName),
					testAccCheckOrgConnectionFolderExists(folderName),
					resource.TestCheckResourceAttr(folderName, "title", title),
					resource.TestCheckResourceAttr(folderName, "parent_id", ""),
				),
			},
			{
				ResourceName: folderName,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccOrgConnectionFolderUpdateConfig(orgHandle, updatedTitle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgName),
					testAccCheckOrgConnectionFolderExists(folderName),
					resource.TestCheckResourceAttr(folderName, "title", updatedTitle),
					resource.TestCheckResourceAttr(folderName, "parent_id", ""),
				),
			},
		},
	})
}

func testAccOrgConnectionFolderConfig(orgHandle, title string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_organization_connection_folder" "folder1" {
	organization = pipes_organization.test.handle
	title  = "%s"
}`, orgHandle, title)
}

func testAccOrgConnectionFolderUpdateConfig(orgHandle, title string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test" {
	handle       = "%s"
	display_name = "Terraform Test Org"
}

resource "pipes_organization_connection_folder" "folder1" {
	organization = pipes_organization.test.handle
	title  = "%s"
}`, orgHandle, title)
}

func testAccCheckOrgConnectionFolderExists(n string) resource.TestCheckFunc {
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

		// Retrieve organization
		org := rs.Primary.Attributes["organization"]

		var r *http.Response
		var err error
		_, r, err = client.APIClient.OrgConnectionFolders.Get(context.Background(), org, connectionFolderId).Execute()
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Connection Folder %s in organization %s not found.\nstatus: %d \nerr: %v", connectionFolderId, org, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccOrgConnectionFolder_Basic testAccCheckOrgConnectionFolderExists %v", err)
			return err
		}

		return nil
	}
}

// testAccCheckOrgConnectionFolderDestroy verifies that a connection folder configured on an organization is destroyed
func testAccCheckOrgConnectionFolderDestroy(s *terraform.State) error {
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
		orgHandle := rs.Primary.Attributes["organization"]

		_, r, err = client.APIClient.OrgConnectionFolders.Get(ctx, orgHandle, connectionFolderId).Execute()
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
