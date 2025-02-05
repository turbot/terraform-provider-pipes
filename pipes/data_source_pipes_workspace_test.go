package pipes

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccWorkspaceDataSource_basic(t *testing.T) {
	dataSourceName := "data.pipes_workspace.workspace_abc"
	workspaceHandle := "abc"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig(workspaceHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "workspace_id", regexp.MustCompile(`^w_[a-z0-9]{20}`)),
					resource.TestCheckResourceAttr(dataSourceName, "handle", workspaceHandle),
					resource.TestCheckResourceAttrSet(dataSourceName, "workspace_state"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig(workspaceHandle string) string {
	return fmt.Sprintf(`
data "pipes_workspace" "workspace_abc" {
	handle = "%s"
}`, workspaceHandle)
}
