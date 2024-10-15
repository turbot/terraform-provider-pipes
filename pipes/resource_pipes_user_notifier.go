package pipes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceUserNotifier() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserNotifierCreate,
		ReadContext:   resourceUserNotifierRead,
		UpdateContext: resourceUserNotifierUpdate,
		DeleteContext: resourceUserNotifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"notifies": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
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
			"state": {
				Type:     schema.TypeString,
				Required: true,
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

func resourceUserNotifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

	s := d.Get("state").(string)
	state, err := pipes.NewNotifierStateFromValue(s)
	if err != nil {
		return diag.Errorf("error parsing state for notifier: %v", err)
	}

	var notifies map[string]interface{}
	if v, ok := d.GetOk("notifies"); ok {
		notifiesString := v.(string)
		err = json.Unmarshal([]byte(notifiesString), &notifies)
		if err != nil {
			return diag.Errorf("error parsing notifies for notifier: %v", err)
		}
	}

	// create request
	client := meta.(*PipesClient)
	req := pipes.CreateNotifierRequest{
		Name:     d.Get("name").(string),
		State:    state,
		Notifies: notifies,
	}

	var userHandle string
	userHandle, r, err = getUserHandler(ctx, client)
	if err != nil {
		return diag.Errorf("error obtaining user handle: %v", decodeResponse(r))
	}

	resp, r, err = client.APIClient.UserNotifiers.Create(ctx, userHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error creating user notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] User notifier created: %s", resp.Name)

	// Set properties
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

	d.SetId(fmt.Sprintf("%s/%s", userHandle, resp.Name))

	return diags
}

func resourceUserNotifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

	var userHandle, notifierName string
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		userHandle = parts[0]
		notifierName = parts[1]
	default:
		return diag.Errorf("error parsing user handle and notifier name from id: %s", d.Id())
	}

	client := meta.(*PipesClient)

	resp, r, err = client.APIClient.UserNotifiers.Get(ctx, userHandle, notifierName).Execute()
	if err != nil {
		return diag.Errorf("error reading user notifier: %v", decodeResponse(r))
	}

	// Set properties
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

	d.SetId(fmt.Sprintf("%s/%s", userHandle, resp.Name))

	return diags
}

func resourceUserNotifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

	var userHandle, notifierName string
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		userHandle = parts[0]
		notifierName = parts[1]
	default:
		return diag.Errorf("error parsing user handle and notifier name from id: %s", d.Id())
	}

	s := d.Get("state").(string)
	state, err := pipes.NewNotifierStateFromValue(s)
	if err != nil {
		return diag.Errorf("error parsing state for notifier: %v", err)
	}

	var notifies map[string]interface{}
	if v, ok := d.GetOk("notifies"); ok {
		notifiesString := v.(string)
		err = json.Unmarshal([]byte(notifiesString), &notifies)
		if err != nil {
			return diag.Errorf("error parsing notifies for notifier: %v", err)
		}
	}

	// create request
	client := meta.(*PipesClient)
	req := pipes.UpdateNotifierRequest{
		Name:     &notifierName,
		State:    state,
		Notifies: &notifies,
	}

	resp, r, err = client.APIClient.UserNotifiers.Update(ctx, userHandle, notifierName).Request(req).Execute()

	// check for errors
	if err != nil {
		return diag.Errorf("error updating user notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] User notifier updated: %s", resp.Name)

	// Set properties
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

	d.SetId(fmt.Sprintf("%s/%s", userHandle, resp.Name))

	return diags
}

func resourceUserNotifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error

	client := meta.(*PipesClient)

	var userHandle, notifierName string
	parts := strings.Split(d.Id(), "/")
	switch len(parts) {
	case 2:
		userHandle = parts[0]
		notifierName = parts[1]
	default:
		return diag.Errorf("error parsing user handle and notifier name from id: %s", d.Id())
	}

	_, r, err = client.APIClient.UserNotifiers.Delete(ctx, userHandle, notifierName).Execute()
	if err != nil {
		return diag.Errorf("resourceUserNotifierDelete error: %v", decodeResponse(r))
	}

	d.SetId("")
	return diags
}
