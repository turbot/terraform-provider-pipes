package pipes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceFlowpipeTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceFlowpipeTriggerCreate,
		ReadContext:   resourceWorkspaceFlowpipeTriggerRead,
		UpdateContext: resourceWorkspaceFlowpipeTriggerUpdate,
		DeleteContext: resourceWorkspaceFlowpipeTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"trigger_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: false,
			},
			"description": {
				Type:     schema.TypeString,
				Required: false,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pipeline": {
				Type:     schema.TypeString,
				Required: true,
			},
			"frequency": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"args": {
				Type:         schema.TypeString,
				Required:     true,
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

func resourceWorkspaceFlowpipeTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var resp pipes.WorkspaceModTrigger
	var err error

	// parse frequency & args - return if error
	var frequency pipes.PipelineFrequency
	var args map[string]interface{}
	err = json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		diags = append(diags, diag.Errorf("error parsing frequency for workspace Flowpipe trigger: %v", d.Get("frequency").(string))...)
	}
	err = json.Unmarshal([]byte(d.Get("args").(string)), &args)
	if err != nil {
		diags = append(diags, diag.Errorf("error parsing args for workspace Flowpipe trigger: %v", d.Get("args").(string))...)
	}
	if len(diags) > 0 {
		return diags
	}

	// Get other fields
	workspaceHandle := d.Get("workspace").(string)
	pipeline := d.Get("pipeline").(string)
	var title, name, description *string
	var state *pipes.TriggerState
	if val, ok := d.GetOk("title"); ok {
		title = val.(*string)
	}
	if val, ok := d.GetOk("name"); ok {
		name = val.(*string)
	}
	if val, ok := d.GetOk("description"); ok {
		description = val.(*string)
	}
	if val, ok := d.GetOk("state"); ok {
		state = val.(*pipes.TriggerState)
		if !state.IsValid() {
			return diag.Errorf("invalid value '%v' for state: valid values are %v", state, pipes.AllowedTriggerStateEnumValues)
		}
	}

	// Create the request
	req := pipes.CreateTriggerRequest{
		Args:        args,
		Description: description,
		Name:        name,
		Pipeline:    pipeline,
		Schedule:    frequency,
		State:       state,
		Title:       title,
	}

	// Create client
	client := meta.(*PipesClient)

	var userHandle string
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeTriggerCreate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeTriggers.Create(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeTriggers.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating workspace Flowpipe trigger: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Trigger: %s created for Pipeline: %s on Workspace: %s", *resp.Id, pipeline, workspaceHandle)

	// Set property values
	d.Set("trigger_id", *resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("workspace", workspaceHandle)
	d.Set("organization", orgHandle)
	d.Set("title", resp.Title)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("pipeline", pipeline)
	d.Set("frequency", FormatJson(resp.Schedule))
	d.Set("args", FormatJson(resp.Args))
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/triggerId" otherwise "workspaceHandle/triggerId"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Id))
	}

	return diags
}

func resourceWorkspaceFlowpipeTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var orgHandle, userHandle, workspaceHandle, triggerNameOrId string
	var isUser = false
	var err error
	var r *http.Response
	var resp pipes.WorkspaceModTrigger

	oldPipeline, newPipeline := d.GetChange("pipeline")
	if oldPipeline.(string) != newPipeline.(string) {
		return diag.Errorf("pipeline is immutable and cannot be changed")
	}
	pipeline := d.Get("pipeline").(string)

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/triggerId" otherwise "workspaceHandle/triggerId"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		triggerNameOrId = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		triggerNameOrId = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<trigger-id> or <org-handle>/<workspace-handle>/<trigger-id>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeTriggerRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeTriggers.Get(ctx, userHandle, workspaceHandle, triggerNameOrId).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeTriggers.Get(ctx, orgHandle, workspaceHandle, triggerNameOrId).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error reading workspace Flowpipe trigger: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Trigger: %s received for Workspace: %s", *resp.Id, workspaceHandle)

	// Set property values
	d.Set("trigger_id", *resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("workspace", workspaceHandle)
	d.Set("organization", orgHandle)
	d.Set("title", resp.Title)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("pipeline", pipeline)
	d.Set("frequency", FormatJson(resp.Schedule))
	d.Set("args", FormatJson(resp.Args))
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/triggerId" otherwise "workspaceHandle/triggerId"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Id))
	}

	return diags
}

func resourceWorkspaceFlowpipeTriggerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var resp pipes.WorkspaceModTrigger
	var err error

	oldPipeline, newPipeline := d.GetChange("pipeline")
	if oldPipeline.(string) != newPipeline.(string) {
		return diag.Errorf("error updating workspace Flowpipe trigger: pipeline is immutable and cannot be changed")
	}
	pipeline := d.Get("pipeline").(string)

	// parse frequency & args - return if error
	var frequency pipes.PipelineFrequency
	var args map[string]interface{}
	err = json.Unmarshal([]byte(d.Get("frequency").(string)), &frequency)
	if err != nil {
		diags = append(diags, diag.Errorf("error parsing frequency for workspace Flowpipe trigger: %v", d.Get("frequency").(string))...)
	}
	err = json.Unmarshal([]byte(d.Get("args").(string)), &args)
	if err != nil {
		diags = append(diags, diag.Errorf("error parsing args for workspace Flowpipe trigger: %v", d.Get("args").(string))...)
	}
	if len(diags) > 0 {
		return diags
	}

	// Get other fields
	triggerId := d.Get("trigger_id").(string)
	workspaceHandle := d.Get("workspace").(string)
	var title, name, description *string
	var state *pipes.TriggerState
	if val, ok := d.GetOk("title"); ok {
		title = val.(*string)
	}
	if val, ok := d.GetOk("name"); ok {
		name = val.(*string)
	}
	if val, ok := d.GetOk("description"); ok {
		description = val.(*string)
	}
	if val, ok := d.GetOk("state"); ok {
		state = val.(*pipes.TriggerState)
		if !state.IsValid() {
			return diag.Errorf("invalid value '%v' for state: valid values are %v", state, pipes.AllowedTriggerStateEnumValues)
		}
	}

	// Create the request
	req := pipes.UpdateTriggerRequest{
		Args:        &args,
		Description: description,
		Name:        name,
		Schedule:    &frequency,
		State:       state,
		Title:       title,
	}

	// Create client
	client := meta.(*PipesClient)

	var userHandle string
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeTriggerUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeTriggers.Update(ctx, userHandle, workspaceHandle, triggerId).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeTriggers.Update(ctx, orgHandle, workspaceHandle, triggerId).Request(req).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error updating workspace Flowpipe trigger: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Trigger: %s updated for Workspace: %s", triggerId, workspaceHandle)

	// Set property values
	d.Set("trigger_id", *resp.Id)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("workspace", workspaceHandle)
	d.Set("organization", orgHandle)
	d.Set("title", resp.Title)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("pipeline", pipeline)
	d.Set("frequency", FormatJson(resp.Schedule))
	d.Set("args", FormatJson(resp.Args))
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/triggerId" otherwise "workspaceHandle/triggerId"
	if userHandle == "" {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Id))
	} else {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Id))
	}

	return diags
}

func resourceWorkspaceFlowpipeTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var orgHandle, userHandle, workspaceHandle, triggerNameOrId string
	var isUser = false

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/triggerId" otherwise "workspaceHandle/triggerId"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		triggerNameOrId = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		triggerNameOrId = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<trigger-id> or <org-handle>/<workspace-handle>/<trigger-id>", d.Id())
	}

	log.Printf("\n[DEBUG] Deleting Trigger: %s for Workspace: %s", triggerNameOrId, workspaceHandle)

	client := meta.(*PipesClient)

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeTriggerDelete.getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceFlowpipeTriggers.Delete(ctx, userHandle, workspaceHandle, triggerNameOrId).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceFlowpipeTriggers.Delete(ctx, orgHandle, workspaceHandle, triggerNameOrId).Execute()
	}

	// Error check
	if err != nil {
		return diag.Errorf("error deleting workspace Flowpipe trigger: %v", decodeResponse(r))
	}

	d.SetId("")

	return diags
}
