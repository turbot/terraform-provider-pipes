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

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspaceRead,
		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"desired_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_volume_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hive": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host": {
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

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var resp pipes.Workspace
	var err error
	var r *http.Response

	var orgHandle, workspaceHandle string

	if v, ok := d.GetOk("organization"); ok {
		orgHandle = v.(string)
	}
	workspaceHandle = d.Get("handle").(string)
	isUser := orgHandle == ""

	client := meta.(*PipesClient)

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("dataSourceWorkspaceRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaces.Get(ctx, userHandle, workspaceHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaces.Get(ctx, orgHandle, workspaceHandle).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error obtaining workspace: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Workspace: %s (%s) received", resp.Handle, resp.Id)

	d.Set("workspace_id", resp.Id)
	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("desired_state", resp.DesiredState)
	d.Set("instance_type", resp.InstanceType)
	d.Set("db_volume_size_bytes", resp.DbVolumeSizeBytes)
	d.Set("database_name", resp.DatabaseName)
	d.Set("hive", resp.Hive)
	d.Set("host", resp.Host)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	if orgHandle != "" {
		d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Handle))
	} else {
		d.SetId(resp.Handle)
	}

	return diags
}
