package pipes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/go-kit/types"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceNotifier() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceNotifierCreate,
		ReadContext:   resourceWorkspaceNotifierRead,
		UpdateContext: resourceWorkspaceNotifierUpdate,
		DeleteContext: resourceWorkspaceNotifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"workspace": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"notifies": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"state": {
				Type:     schema.TypeString,
				Required: true,
			},
			"notifier_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"precedence": {
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

func resourceWorkspaceNotifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var resp pipes.Notifier
	var err error
	var tfId string

	workspaceHandle := d.Get("workspace").(string)
	notifierName := d.Get("name").(string)
	isUser, orgHandle := isUserConnection(d)

	s := d.Get("state").(string)
	state, err := pipes.NewNotifierStateFromValue(s)
	if err != nil {
		return diag.Errorf("error parsing state for notifier: %v", err)
	}

	var notifies []map[string]interface{}
	if v, ok := d.GetOk("notifies"); ok {
		notifiesString := v.(string)
		err = json.Unmarshal([]byte(notifiesString), &notifies)
		if err != nil {
			return diag.Errorf("error parsing notifies for notifier: %v", err)
		}
	}

	client := meta.(*PipesClient)
	req := pipes.CreateNotifierRequest{
		Name:     notifierName,
		State:    state,
		Notifies: notifies,
	}

	if isUser {
		var userHandle string
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("error obtaining user handle: %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceNotifiers.Create(ctx, userHandle, workspaceHandle).Request(req).Execute()
		tfId = fmt.Sprintf("%s/%s", workspaceHandle, notifierName)
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceNotifiers.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
		tfId = fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, notifierName)
	}
	if err != nil {
		return diag.Errorf("error creating workspace notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Workspace notifier created: %s", resp.Name)

	// Set properties
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	d.Set("name", resp.Name)
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("updated_at", resp.UpdatedAt)
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	d.SetId(tfId)

	return diags
}

func resourceWorkspaceNotifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier
	var orgHandle, userHandle, workspaceHandle, notifierName string
	var isUser = false

	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		notifierName = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		notifierName = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<notifier-name> or <org-handle>/<workspace-handle>/<notifier-name>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceNotifierRead.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceNotifiers.Get(ctx, userHandle, workspaceHandle, notifierName).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceNotifiers.Get(ctx, orgHandle, workspaceHandle, notifierName).Execute()
	}
	if err != nil {
		return diag.Errorf("error reading workspace notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Notifier: %s received for Workspace: %s", resp.Name, workspaceHandle)

	// Set properties
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	d.Set("name", resp.Name)
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("updated_at", resp.UpdatedAt)
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	d.SetId(d.Id())
	return diags
}

func resourceWorkspaceNotifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier
	var orgHandle, userHandle, workspaceHandle, notifierName string
	var isUser = false

	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		notifierName = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		notifierName = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<notifier-name> or <org-handle>/<workspace-handle>/<notifier-name>", d.Id())
	}

	s := d.Get("state").(string)
	state, err := pipes.NewNotifierStateFromValue(s)
	if err != nil {
		return diag.Errorf("error parsing state for notifier: %v", err)
	}

	var notifies []map[string]interface{}
	if v, ok := d.GetOk("notifies"); ok {
		notifiesString := v.(string)
		err = json.Unmarshal([]byte(notifiesString), &notifies)
		if err != nil {
			return diag.Errorf("error parsing notifies for notifier: %v", err)
		}
	}

	newNotifierName := notifierName
	_, newName := d.GetChange("name")
	if newName != nil && newName.(string) != notifierName {
		newNotifierName = newName.(string)
	}

	client := meta.(*PipesClient)
	req := pipes.UpdateNotifierRequest{
		Name:     types.String(newNotifierName),
		State:    state,
		Notifies: &notifies,
	}

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceNotifierUpdate.getUserHandler error  %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceNotifiers.Update(ctx, userHandle, workspaceHandle, notifierName).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceNotifiers.Update(ctx, orgHandle, workspaceHandle, notifierName).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("error updating workspace notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Notifier: %s updated for Workspace: %s", resp.Name, workspaceHandle)

	// Set properties
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	d.Set("name", resp.Name)
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("created_at", resp.CreatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	d.Set("updated_at", resp.UpdatedAt)
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, resp.Name))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, resp.Name))
	}
	return diags
}

func resourceWorkspaceNotifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var orgHandle, userHandle, workspaceHandle, notifierName string
	var isUser = false

	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		isUser = true
		workspaceHandle = parts[0]
		notifierName = parts[1]
	case 3:
		orgHandle = parts[0]
		workspaceHandle = parts[1]
		notifierName = parts[2]
	default:
		return diag.Errorf("unexpected format for ID (%q), expected <workspace-handle>/<notifier-name> or <org-handle>/<workspace-handle>/<notifier-name>", d.Id())
	}

	client := meta.(*PipesClient)

	if isUser {
		userHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceNotifierDelete.getUserHandler error  %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceNotifiers.Delete(ctx, userHandle, workspaceHandle, notifierName).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceNotifiers.Delete(ctx, orgHandle, workspaceHandle, notifierName).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceNotifierDelete error: %v", decodeResponse(r))
	}

	d.SetId("")
	return diags
}
