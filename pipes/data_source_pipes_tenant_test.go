package pipes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenantDataSource_basic(t *testing.T) {
	dataSourceName := "data.pipes_tenant.primary_tenant"
	tenantHandle := PipesTenantHandle

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfig(tenantHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tenant_id", PipesTenantId),
					resource.TestCheckResourceAttr(dataSourceName, "id", PipesTenantId),
					resource.TestCheckResourceAttr(dataSourceName, "handle", tenantHandle),
					resource.TestCheckResourceAttr(dataSourceName, "state", "created"),
				),
			},
		},
	})
}

func testAccTenantDataSourceConfig(tenantHandle string) string {
	return fmt.Sprintf(`
data "pipes_tenant" "primary_tenant" {
	handle = "%s"
}`, tenantHandle)
}
