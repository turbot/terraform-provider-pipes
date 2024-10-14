package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceFlowpipeModVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceFlowpipeModVariableCreate,
		ReadContext:   resourceWorkspaceFlowpipeModVariableRead,
		UpdateContext: resourceWorkspaceFlowpipeModVariableUpdate,
		DeleteContext: resourceWorkspaceFlowpipeModVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_mod_variable_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"default_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": {
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
				Optional: true,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"mod_alias": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9_]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
		},
	}
}

func resourceWorkspaceFlowpipeModVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.WorkspaceModVariable

	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("mod_alias").(string)
	variableName := d.Get("name").(string)
	settingRaw := d.Get("setting_value")
	setting, err := JSONStringToInterface(settingRaw.(string))
	if err != nil {
		return diag.Errorf("error parsing setting for workspace mod variable : %v", setting)
	}

	// Create request
	req := pipes.CreateWorkspaceModVariableSettingRequest{Name: variableName, Setting: setting}
	log.Printf("\n[DEBUG] Request Setting : %v \n", req.Setting)

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModVariableCreate.getUserHandle error  %v", decodeResponse(r))
		}
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			resp, r, err = client.APIClient.UserWorkspaceFlowpipeModVariables.CreateSetting(ctx, userHandle, workspaceHandle, modAlias).Request(req).Execute()
			if err != nil {
				return resource.RetryableError(err)
			}
			return nil
		})
	} else {
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			resp, r, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.CreateSetting(ctx, orgHandle, workspaceHandle, modAlias).Request(req).Execute()
			if err != nil {
				return resource.RetryableError(err)
			}
			return nil
		})
	}

	// Error check
	if err != nil {
		return diag.Errorf("error creating setting for workspace mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Setting created for variable: %s of Flowpipe mod: %s in workspace: %s", variableName, modAlias, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceFlowpipeModVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias, variableName string
	var isUser = false
	var err error
	var r *http.Response
	var resp pipes.WorkspaceModVariable

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/modAlias/variableName" otherwise "workspaceHandle/modAlias/variableName"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 3:
		isUser = true
		workspaceHandle = parts[0]
		modAlias = parts[1]
		variableName = parts[2]
	case 4:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		modAlias = parts[2]
		variableName = parts[3]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<mod-alias> or <org-handle>/<workspace-handle>/<mod-alias>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModVariableRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeModVariables.GetSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.GetSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error fetching setting for workspace Flowpipe mod variable: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Variable: %s received for Flowpipe mod: %s in Workspace: %s", variableName, modAlias, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceFlowpipeModVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.WorkspaceModVariable

	// TODO: #query should we ensure that workspaceHandle, modAlias and variableName are not changed? - Should we pull from d.Id() like a read?
	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("mod_alias").(string)
	variableName := d.Get("name").(string)
	settingRaw := d.Get("setting_value")
	setting, err := JSONStringToInterface(settingRaw.(string))
	if err != nil {
		return diag.Errorf("error parsing setting for workspace mod variable : %v", setting)
	}

	// Create request
	req := pipes.UpdateWorkspaceModVariableSettingRequest{Setting: setting}

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModVariableUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeModVariables.UpdateSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.UpdateSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Request(req).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error updating setting for workspace Flowpipe mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Setting updated for Variable: %s of Flowpipe mod: %s in Workspace: %s", variableName, modAlias, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_variable_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("workspace_handle", workspaceHandle)
	d.Set("mod_alias", modAlias)
	d.Set("organization", orgHandle)
	d.Set("default_value", FormatJson(resp.ValueDefault))
	d.Set("setting_value", FormatJson(resp.ValueSetting))
	d.Set("value", FormatJson(resp.Value))

	// If the mod variable belongs to a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias/VariableName" otherwise "WorkspaceHandle/ModAlias/VariableName"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s/%s", workspaceHandle, modAlias, variableName))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s/%s", orgHandle, workspaceHandle, modAlias, variableName))
	}

	return diags
}

func resourceWorkspaceFlowpipeModVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias, variableName string
	var isUser = false
	var err error
	var r *http.Response

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/modAlias/variableName" otherwise "workspaceHandle/modAlias/variableName"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 3:
		isUser = true
		workspaceHandle = parts[0]
		modAlias = parts[1]
		variableName = parts[2]
	case 4:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		modAlias = parts[2]
		variableName = parts[3]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<mod-alias> or <org-handle>/<workspace-handle>/<mod-alias>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModVariableRead.getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceFlowpipeModVariables.DeleteSetting(ctx, userHandle, workspaceHandle, modAlias, variableName).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceFlowpipeModVariables.DeleteSetting(ctx, orgHandle, workspaceHandle, modAlias, variableName).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error deleting setting for workspace Flowpipe mod variable : %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Setting deleted for Variable: %s of Flowpipe mod: %s in Workspace: %s", variableName, modAlias, workspaceHandle)

	d.SetId("")

	return diags
}
