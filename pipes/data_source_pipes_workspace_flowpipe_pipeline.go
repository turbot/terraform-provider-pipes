package pipes

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/turbot/pipes-sdk-go"
)

func dataSourceWorkspaceFlowpipePipeline() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspaceFlowpipeModPipelineRead,
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"pipeline": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"args": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"desired_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSourceWorkspaceFlowpipeModPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var resp pipes.Pipeline
	var r *http.Response
	var err error

	workspace := d.Get("workspace").(string)
	pipelineId := d.Get("pipeline_id").(string)

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("dataSourceWorkspaceFlowpipeModPipelineRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipePipelines.Get(ctx, userHandle, workspace, pipelineId).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipePipelines.Get(ctx, orgHandle, workspace, pipelineId).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error reading workspace Flowpipe pipeline: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Pipeline: %s received for Workspace: %s", resp.Id, workspace)

	// Set properties
	d.Set("args", FormatJson(resp.Args))
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("desired_state", resp.DesiredState)
	d.Set("frequency", FormatJson(resp.Frequency))
	d.Set("pipeline_id", resp.Id)
	d.Set("pipeline", resp.Pipeline)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("title", resp.Title)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("workspace", workspace)
	d.Set("workspace_id", resp.WorkspaceId)
	d.SetId(resp.Id)

	return diags
}
