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

func TestAccTenantConnectionFolderPermission_Basic(t *testing.T) {
	orgResource1 := "pipes_organization.test_org_1"
	orgResource2 := "pipes_organization.test_org_2"
	connFolderResource := "pipes_tenant_connection_folder.folder_1"
	permissionResource := "pipes_tenant_connection_folder_permission.permission_1"
	orgHandle1 := "org" + randomString(5)
	orgHandle2 := "org" + randomString(6)
	folderTitle := "My Tenant Level Connection Folder"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTenantConnectionFolderPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConnectionFolderPermissionConfig(orgHandle1, orgHandle2, folderTitle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource1),
					testAccCheckOrganizationExists(orgResource2),
					testAccCheckTenantConnectionFolderExists(connFolderResource),
					testAccCheckTenantConnectionFolderPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle1),
					testAccCheckTenantConnectionFolderAccess(connFolderResource, orgHandle1, orgHandle2),
				),
			},
			{
				ResourceName: permissionResource,
				ImportState:  true,
				// ImportStateVerify: true,
			},
			{
				Config: testAccTenantConnectionFolderPermissionUpdateConfig(orgHandle1, orgHandle2, folderTitle),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(orgResource1),
					testAccCheckOrganizationExists(orgResource2),
					testAccCheckTenantConnectionFolderExists(connFolderResource),
					testAccCheckTenantConnectionFolderPermissionExists(permissionResource),
					resource.TestCheckResourceAttr(permissionResource, "identity_handle", orgHandle2),
					testAccCheckTenantConnectionFolderAccess(connFolderResource, orgHandle2, orgHandle1),
				),
			},
		},
	})
}

// configs
func testAccTenantConnectionFolderPermissionConfig(orgHandle1, orgHandle2, folderTitle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org_1" {
	handle       = "%s"
	display_name = "Terraform Test Org 1"
}

resource "pipes_organization" "test_org_2" {
	handle       = "%s"
	display_name = "Terraform Test Org 2"
}

resource "pipes_tenant_connection_folder" "folder_1" {
	title = "%s"
}

resource "pipes_tenant_connection_folder_permission" "permission_1" {
	connection_folder_id = pipes_tenant_connection_folder.folder_1.connection_folder_id
	identity_handle      = pipes_organization.test_org_1.handle
}`, orgHandle1, orgHandle2, folderTitle)
}

func testAccTenantConnectionFolderPermissionUpdateConfig(orgHandle1, orgHandle2, folderTitle string) string {
	return fmt.Sprintf(`
resource "pipes_organization" "test_org_1" {
	handle       = "%s"
	display_name = "Terraform Test Org 1"
}

resource "pipes_organization" "test_org_2" {
	handle       = "%s"
	display_name = "Terraform Test Org 2"
}

resource "pipes_tenant_connection_folder" "folder_1" {
	title = "%s"
}

resource "pipes_tenant_connection_folder_permission" "permission_1" {
	connection_folder_id = pipes_tenant_connection_folder.folder_1.connection_folder_id
	identity_handle      = pipes_organization.test_org_2.handle
}`, orgHandle1, orgHandle2, folderTitle)
}

// testAccCheckTenantConnectionFolderPermissionDestroy verifies the connection permission has been destroyed
func testAccCheckTenantConnectionFolderPermissionDestroy(s *terraform.State) error {
	var r *http.Response
	var err error
	ctx := context.Background()

	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*PipesClient)

	// loop through the resources in state, verifying each connection is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pipes_tenant_connection_folder_permission" {
			continue
		}

		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]
		permissionId := rs.Primary.Attributes["permission_id"]

		_, r, err = client.APIClient.TenantConnectionFolders.GetPermission(ctx, connectionFolderId, permissionId).Execute()
		if err == nil {
			return fmt.Errorf("Permission on connection folder %s still exists.", connectionFolderId)
		}

		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if r.StatusCode != 404 {
			log.Printf("[INFO] TestAccTenantConnectionFolderPermission_Basic testAccCheckTenantConnectionFolderPermissionDestroy %v", err)
			return fmt.Errorf("status: %d \nerr: %v", r.StatusCode, r.Body)
		}

	}

	return nil
}

func testAccCheckTenantConnectionFolderPermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		connectionFolderId := rs.Primary.Attributes["connection_folder_id"]
		permissionId := rs.Primary.Attributes["permission_id"]

		client := testAccProvider.Meta().(*PipesClient)

		var r *http.Response
		var err error

		_, r, err = client.APIClient.TenantConnectionFolders.GetPermission(context.Background(), connectionFolderId, permissionId).Execute()
		// If the error is equivalent to 404 not found, the connection is destroyed.
		// Otherwise return the error
		if err != nil {
			if r.StatusCode != 404 {
				return fmt.Errorf("Permission on connection folder %s in tenant not found.\nstatus: %d \nerr: %v", connectionFolderId, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnectionFolderPermission_Basic testAccCheckTenantConnectionFolderPermissionExists %v", err)
			return err
		}
		return nil
	}
}

func testAccCheckTenantConnectionFolderAccess(n, orgAvailable, orgNotAvailable string) resource.TestCheckFunc {
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

		// The connection folder should be accessible by the organization with handle `orgAvailable`
		_, r, err = client.APIClient.OrgConnectionFolders.Get(context.Background(), orgAvailable, connectionFolderId).Execute()
		// If there's an error and its a not found error, it means the connection is not available to the organization, fail the test
		if err != nil {
			if r.StatusCode == 404 {
				return fmt.Errorf("Connection Folder %s in organization %s not found.\nstatus: %d \nerr: %v", connectionFolderId, orgAvailable, r.StatusCode, r.Body)
			}
			log.Printf("[INFO] TestAccTenantConnectionFolderPermission_Basic testAccCheckTenantConnectionAccess %v", err)
			return err
		}

		// The connection folder should not be accessible by the organization with handle `orgNotAvailable`
		_, r, err = client.APIClient.OrgConnectionFolders.Get(context.Background(), orgNotAvailable, connectionFolderId).Execute()
		// If there's no error here, it means the connection has bern returned which is not expected, fail the test
		if err == nil {
			return fmt.Errorf("Connection Folder %s should not be available to organization %s.\nstatus: %d \nerr: %v", connectionFolderId, orgNotAvailable, r.StatusCode, r.Body)
		}

		return nil
	}
}
