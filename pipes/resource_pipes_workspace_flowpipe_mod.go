package pipes

import (
	"context"
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

func resourceWorkspaceFlowpipeMod() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceFlowpipeModInstall,
		ReadContext:   resourceWorkspaceFlowpipeModRead,
		UpdateContext: resourceWorkspaceFlowpipeModUpdate,
		DeleteContext: resourceWorkspaceFlowpipeModUninstall,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workspace_mod_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Optional: false,
				Computed: true,
			},
			"constraint": {
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
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"installed_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
		},
	}
}

func resourceWorkspaceFlowpipeModInstall(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.WorkspaceMod
	var constraint *string

	workspaceHandle := d.Get("workspace_handle").(string)
	path := d.Get("path").(string)
	if val, ok := d.GetOk("constraint"); ok {
		constraint = val.(*string)
	}

	// Create the request
	req := pipes.CreateWorkspaceModRequest{
		Path:       path,
		Constraint: constraint,
	}

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceConnectionCreate. getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeMods.Install(ctx, userHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeMods.Install(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error creating workspace Flowpipe mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Flowpipe Mod: %s installed for Workspace: %s", *resp.Path, workspaceHandle)
	log.Printf("\n[DEBUG] Flowpipe Mod Alias: %s", *resp.Alias)

	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("path", resp.Path)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias" otherwise "WorkspaceHandle/ModAlias"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceFlowpipeModRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias string
	var isUser = false
	var resp pipes.WorkspaceMod
	var err error
	var r *http.Response

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/modAlias" otherwise "workspaceHandle/modAlias"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		modAlias = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		modAlias = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<mod-alias> or <org-handle>/<workspace-handle>/<mod-alias>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeMods.Get(ctx, userHandle, workspaceHandle, modAlias).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeMods.Get(ctx, orgHandle, workspaceHandle, modAlias).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error getting workspace mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Flowpipe Mod: %s received for Workspace: %s", *resp.Path, workspaceHandle)

	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("path", resp.Path)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias" otherwise "WorkspaceHandle/ModAlias"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceFlowpipeModUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var resp pipes.WorkspaceMod
	var err error

	// TODO: #query should we ensure that workspaceHandle/modAlias are not changed? - Should we pull from d.Id() like a read?
	workspaceHandle := d.Get("workspace_handle").(string)
	modAlias := d.Get("alias").(string)
	constraint := d.Get("constraint").(string)

	req := pipes.UpdateWorkspaceModRequest{
		Constraint: &constraint,
	}

	client := meta.(*PipesClient)

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceFlowpipeMods.Update(ctx, userHandle, workspaceHandle, modAlias).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceFlowpipeMods.Update(ctx, orgHandle, workspaceHandle, modAlias).Request(req).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error updating workspace Flowpipe mod: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Flowpipe Mod: %s updated for Workspace: %s", *resp.Path, workspaceHandle)

	// Set property values
	// Set property values
	d.Set("workspace_mod_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("constraint", resp.Constraint)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("alias", resp.Alias)
	d.Set("installed_version", resp.InstalledVersion)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("path", resp.Path)
	d.Set("organization", orgHandle)
	d.Set("workspace_handle", workspaceHandle)

	// If mod is installed for a workspace inside an Organization the id will be of the
	// format "OrganizationHandle/WorkspaceHandle/ModAlias" otherwise "WorkspaceHandle/ModAlias"
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Alias))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Alias))
	}

	return diags
}

func resourceWorkspaceFlowpipeModUninstall(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, modAlias string
	var isUser = false
	var err error
	var r *http.Response

	// If a trigger is created for a workspace inside an organization the id will be of the
	// format "orgHandle/workspaceHandle/modAlias" otherwise "workspaceHandle/modAlias"
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		modAlias = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		modAlias = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<mod-alias> or <org-handle>/<workspace-handle>/<mod-alias>", d.Id())
	}

	log.Printf("\n[DEBUG] Uninstalling Flowpipe Mod: %s for Workspace: %s", modAlias, workspaceHandle)

	client := meta.(*PipesClient)

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceFlowpipeModUninstall.getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceFlowpipeMods.Uninstall(ctx, userHandle, workspaceHandle, modAlias).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceFlowpipeMods.Uninstall(ctx, orgHandle, workspaceHandle, modAlias).Execute()
	}

	// Check for errors
	if err != nil {
		return diag.Errorf("error uninstalling workspace Flowpipe mod: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
