package pipes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenantIntegrationDataSource_Basic(t *testing.T) {
	dataSourceName := "data.pipes_integration.test"
	handle := "pipes-email"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantIntegrationDataSourceConfig(handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "handle", handle),
					resource.TestCheckResourceAttr(dataSourceName, "type", "email"),
				),
			},
		},
	})
}

func testAccTenantIntegrationDataSourceConfig(handle string) string {
	return fmt.Sprintf(`
data "pipes_tenant_integration" "test" {
	handle = "%s"
}`, handle)
}

func TestAccUserIntegrationDataSource_Basic(t *testing.T) {
	dataSourceName := "data.pipes_integration.test"
	handle := "pipes-email"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserIntegrationDataSourceConfig(handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "handle", handle),
					resource.TestCheckResourceAttr(dataSourceName, "type", "email"),
				),
			},
		},
	})
}

func testAccUserIntegrationDataSourceConfig(handle string) string {
	return fmt.Sprintf(`
data "pipes_user_integration" "test" {
	handle = "%s"
}`, handle)
}
