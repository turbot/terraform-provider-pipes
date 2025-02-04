package pipes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceTenantNotifier() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantNotifierCreate,
		ReadContext:   resourceTenantNotifierRead,
		UpdateContext: resourceTenantNotifierUpdate,
		DeleteContext: resourceTenantNotifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"notifier_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
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
			"precedence": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceTenantNotifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

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

	notifierName := d.Get("name").(string)

	// create request
	client := meta.(*PipesClient)
	req := pipes.CreateNotifierRequest{
		Name:     notifierName,
		State:    state,
		Notifies: notifies,
	}

	resp, r, err = client.APIClient.TenantNotifiers.Create(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error creating tenant notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Tenant notifier created: %s", resp.Name)

	// Set properties
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("name", resp.Name)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
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

	d.SetId(resp.Name)

	return diags
}

func resourceTenantNotifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

	notifierName := d.Id()

	client := meta.(*PipesClient)

	resp, r, err = client.APIClient.TenantNotifiers.Get(ctx, notifierName).Execute()
	if err != nil {
		return diag.Errorf("error reading tenant notifier: %v", decodeResponse(r))
	}

	// Set properties
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("name", resp.Name)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
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

	d.SetId(resp.Name)

	return diags
}

func resourceTenantNotifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error
	var resp pipes.Notifier

	oldName, newName := d.GetChange("name")
	oldNotifierName := oldName.(string)
	newNotifierName := newName.(string)

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

	// create request
	client := meta.(*PipesClient)
	req := pipes.UpdateNotifierRequest{
		Name:     &newNotifierName,
		State:    state,
		Notifies: &notifies,
	}

	resp, r, err = client.APIClient.TenantNotifiers.Update(ctx, oldNotifierName).Request(req).Execute()
	// check for errors
	if err != nil {
		return diag.Errorf("error updating tenant notifier: %v", decodeResponse(r))
	}
	log.Printf("\n[DEBUG] Tenant notifier updated: %s", resp.Name)

	// Set properties
	d.Set("notifier_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("name", resp.Name)
	d.Set("notifies", FormatJson(resp.Notifies))
	d.Set("precedence", resp.Precedence)
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

	d.SetId(resp.Name)

	return diags
}

func resourceTenantNotifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var r *http.Response
	var err error

	client := meta.(*PipesClient)
	notifierName := d.Id()

	_, r, err = client.APIClient.TenantNotifiers.Delete(ctx, notifierName).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantNotifierDelete error: %v", decodeResponse(r))
	}

	d.SetId("")
	return diags
}
