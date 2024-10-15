package pipes

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccUserWorkspaceIntegrationDataSource_Basic(t *testing.T) {
	dataSourceName := "data.pipes_integration.test"
	workspaceHandle := "abc"
	handle := "pipes-email"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccUserWorkspaceIntegrationDataSourceConfig(workspaceHandle, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "handle", handle),
					resource.TestCheckResourceAttr(dataSourceName, "type", "email"),
				),
			},
		},
	})
}

func testAccUserWorkspaceIntegrationDataSourceConfig(workspaceHandle string, handle string) string {
	return fmt.Sprintf(`
data "pipes_integration" "test" {
	workspace = "%s"
	handle = "%s"
}`, workspaceHandle, handle)
}
