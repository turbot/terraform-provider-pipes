package pipes

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/turbot/pipes-sdk-go"
)

func resourceTenantServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantServiceAccountCreate,
		ReadContext:   resourceTenantServiceAccountRead,
		UpdateContext: resourceTenantServiceAccountUpdate,
		DeleteContext: resourceTenantServiceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
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

func resourceTenantServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	req := pipes.CreateServiceAccountUserRequest{}
	if v, ok := d.GetOk("title"); ok {
		val := v.(string)
		req.SetTitle(val)
	}
	if v, ok := d.GetOk("description"); ok {
		val := v.(string)
		req.SetDescription(val)
	}

	resp, r, err := client.APIClient.TenantServiceAccounts.Create(ctx).Body(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantServiceAccountCreate. Create tenant service account error: %v", decodeResponse(r))
	}

	setTenantServiceAccountFields(d, &resp)
	// For tenant-level service accounts the identifier is globally unique; use it as ID
	d.SetId(resp.Id)

	return diags
}

func resourceTenantServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	id := d.Id()

	resp, r, err := client.APIClient.TenantServiceAccounts.Get(ctx, id).Execute()
	if err != nil {
		if r != nil && r.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceTenantServiceAccountRead. Get tenant service account error: %v", decodeResponse(r))
	}

	setTenantServiceAccountFields(d, &resp)
	d.SetId(resp.Id)

	return diags
}

func resourceTenantServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	id := d.Id()

	req := pipes.UpdateServiceAccountUserRequest{}
	var hasUpdate bool
	if d.HasChange("title") {
		v := d.Get("title").(string)
		req.SetTitle(v)
		hasUpdate = true
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		req.SetDescription(v)
		hasUpdate = true
	}
	if hasUpdate {
		resp, r, err := client.APIClient.TenantServiceAccounts.Update(ctx, id).Body(req).Execute()
		if err != nil {
			return diag.Errorf("resourceTenantServiceAccountUpdate. Update tenant service account error: %v", decodeResponse(r))
		}
		setTenantServiceAccountFields(d, &resp)
	}

	return diags
}

func resourceTenantServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	id := d.Id()
	r, err := client.APIClient.TenantServiceAccounts.Delete(ctx, id).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantServiceAccountDelete. Delete tenant service account error: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}

func setTenantServiceAccountFields(d *schema.ResourceData, resp *pipes.User) {
	// Common fields from User model
	d.Set("service_account_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	if resp.Title != nil {
		d.Set("title", resp.GetTitle())
	}
	if resp.Description != nil {
		d.Set("description", resp.GetDescription())
	}
	d.Set("created_at", resp.CreatedAt)
	if resp.UpdatedAt != nil {
		d.Set("updated_at", resp.GetUpdatedAt())
	}
	d.Set("version_id", resp.VersionId)
}
