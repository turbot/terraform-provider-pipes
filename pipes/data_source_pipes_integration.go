package pipes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/turbot/pipes-sdk-go"
	"log"
	"net/http"
)

func dataSourceIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIntegrationRead,
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"workspace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeString,
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

func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var resp pipes.Integration
	var r *http.Response
	var err error

	client := meta.(*PipesClient)

	orgHandle := d.Get("organization").(string)
	workspaceHandle := d.Get("workspace").(string)
	integrationHandle := d.Get("handle").(string)

	var userHandle, tfId string
	isUser := orgHandle == ""

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("dataSourceIntegrationRead.getUserHandler error  %v", decodeResponse(r))
		}
	}

	switch {
	case isUser && workspaceHandle == "":
		resp, r, err = client.APIClient.UserIntegrations.Get(ctx, userHandle, integrationHandle).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s", userHandle, resp.Handle)
		}
	case isUser && workspaceHandle != "":
		resp, r, err = client.APIClient.UserWorkspaceIntegrations.Get(ctx, userHandle, workspaceHandle, integrationHandle).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s/%s", userHandle, workspaceHandle, resp.Handle)
		}
	case !isUser && workspaceHandle == "":
		resp, r, err = client.APIClient.OrgIntegrations.Get(ctx, orgHandle, integrationHandle).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s", orgHandle, resp.Handle)
		}
	case !isUser && workspaceHandle != "":
		resp, r, err = client.APIClient.OrgWorkspaceIntegrations.Get(ctx, orgHandle, workspaceHandle, integrationHandle).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Handle)
		}
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error obtaining integration: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Integration: %s received", resp.Id)

	// Convert config to string
	configString, err := mapToJSONString(resp.GetConfig())
	if err != nil {
		return diag.Errorf("resourceTenantIntegrationRead. Error converting config to string: %v", err)
	}

	// Set properties
	d.Set("integration_id", resp.Id)
	d.Set("handle", resp.Handle)
	d.Set("identity_id", resp.IdentityId)
	d.Set("tenant_id", resp.TenantId)
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

	d.SetId(tfId)
	return diags
}
