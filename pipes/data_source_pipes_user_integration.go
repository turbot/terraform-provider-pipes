package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/turbot/pipes-sdk-go"
)

func dataSourceUserIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserIntegrationRead,
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"github_installation_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"pipeline_id": {
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
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceUserIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var resp pipes.Integration
	var r *http.Response
	var err error

	client := meta.(*PipesClient)

	integrationHandle := d.Get("handle").(string)
	var userHandle string
	userHandle, r, err = getUserHandler(ctx, client)
	if err != nil {
		return diag.Errorf("error obtaining user handle: %v", decodeResponse(r))
	}

	resp, r, err = client.APIClient.UserIntegrations.Get(ctx, userHandle, integrationHandle).Execute()
	if err != nil {
		return diag.Errorf("error obtaining integration: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Integration: %s received", resp.Id)

	// Convert config to string
	configString, err := mapToJSONString(resp.GetConfig())
	if err != nil {
		return diag.Errorf("dataSourceUserIntegrationRead. Error converting config to string: %v", err)
	}

	// Set properties
	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("config", configString)
	d.Set("github_installation_id", resp.GithubInstallationId)
	d.Set("pipeline_id", resp.PipelineId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	d.SetId(fmt.Sprintf("%s/%s", userHandle, integrationHandle))
	return diags
}
