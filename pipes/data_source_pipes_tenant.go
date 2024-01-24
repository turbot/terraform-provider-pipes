package pipes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTenantRead,
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"handle", "tenant_id"},
			},
			"tenant_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"handle", "tenant_id"},
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceTenantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var handle string
	handle = d.Get("handle").(string)
	if strings.TrimSpace(handle) == "" {
		handle = d.Get("tenant_id").(string)
	}
	if strings.TrimSpace(handle) == "" {
		return diags
	}

	resp, r, err := client.APIClient.Tenants.Get(ctx, handle).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("%v", decodeResponse(r)))
	}

	if err := d.Set("handle", resp.Handle); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_id", resp.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", resp.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("avatar_url", resp.AvatarUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", resp.State); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", resp.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", resp.UpdatedAt); err != nil {
		return diag.FromErr(err)
	}
	if resp.CreatedBy != nil {
		if err := d.Set("created_by", resp.CreatedBy.Handle); err != nil {
			return diag.FromErr(err)
		}
	}
	if resp.UpdatedBy != nil {
		if err := d.Set("updated_by", resp.UpdatedBy.Handle); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("version_id", resp.VersionId); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Id)

	return diags
}
