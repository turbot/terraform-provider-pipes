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

func dataSourceWorkspaceFlowpipePipeline() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspaceFlowpipeModPipelineRead,
		Schema: map[string]*schema.Schema{
			"workspace_mod_pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"workspace_mod_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"params": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"steps": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"triggers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_process_id": {
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
		},
	}
}

func dataSourceWorkspaceFlowpipeModPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var resp pipes.WorkspaceModPipeline
	var r *http.Response
	var err error

	workspace := d.Get("workspace").(string)
	pipelineId := d.Get("workspace_mod_pipeline_id").(string)
	var tfId string

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("dataSourceWorkspaceFlowpipeModPipelineRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipePipelines.Get(ctx, userHandle, workspace, pipelineId).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s", workspace, pipelineId)
		}
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipePipelines.Get(ctx, orgHandle, workspace, pipelineId).Execute()
		if err == nil {
			tfId = fmt.Sprintf("%s/%s/%s", orgHandle, workspace, pipelineId)
		}
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error reading workspace Flowpipe pipeline: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Pipeline: %s received for Workspace: %s", *resp.Id, workspace)

	// Set properties
	d.Set("workspace_mod_pipeline_id", FormatJson(resp.Id))
	d.Set("workspace_mod_id", resp.WorkspaceModId)
	d.Set("name", resp.Name)
	d.Set("title", resp.Title)
	d.Set("description", resp.Description)
	d.Set("params", FormatJson(resp.Params))
	d.Set("steps", FormatJson(resp.Steps))
	d.Set("tags", FormatJson(resp.Tags))
	d.Set("triggers", FormatJson(resp.Triggers))
	d.Set("last_process_id", resp.LastProcessId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.SetId(tfId)

	return diags
}
