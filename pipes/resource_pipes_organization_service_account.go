package pipes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/turbot/pipes-sdk-go"
)

func resourceOrganizationServiceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationServiceAccountCreate,
		ReadContext:   resourceOrganizationServiceAccountRead,
		UpdateContext: resourceOrganizationServiceAccountUpdate,
		DeleteContext: resourceOrganizationServiceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
				// Expect import ID in the form <org_handle>/<service_account_identifier>
				id := d.Id()
				var orgHandle, saId string
				for i := 0; i < len(id); i++ {
					if id[i] == '/' {
						orgHandle = id[:i]
						saId = id[i+1:]
						break
					}
				}
				if orgHandle == "" || saId == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected <org_handle>/<service_account_identifier>", id)
				}
				if err := d.Set("organization_handle", orgHandle); err != nil {
					return nil, err
				}
				// Use service account identifier as the Terraform ID
				d.SetId(saId)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"organization_handle": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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

func resourceOrganizationServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	orgHandle := d.Get("organization_handle").(string)

	req := pipes.CreateServiceAccountUserRequest{}
	if v, ok := d.GetOk("title"); ok {
		req.SetTitle(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.SetDescription(v.(string))
	}

	resp, r, err := client.APIClient.OrgServiceAccounts.Create(ctx, orgHandle).Body(req).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationServiceAccountCreate. Create org service account error: %v", decodeResponse(r))
	}

	setOrganizationServiceAccountFields(d, &resp)
	// Set Terraform ID to the service account identifier only
	d.SetId(resp.Id)

	return diags
}

func resourceOrganizationServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	orgHandle := d.Get("organization_handle").(string)
	saId := d.Id()

	resp, r, err := client.APIClient.OrgServiceAccounts.Get(ctx, orgHandle, saId).Execute()
	if err != nil {
		if r != nil && r.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceOrganizationServiceAccountRead. Get org service account error: %v", decodeResponse(r))
	}

	setOrganizationServiceAccountFields(d, &resp)
	d.SetId(resp.Id)

	return diags
}

func resourceOrganizationServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	orgHandle := d.Get("organization_handle").(string)
	saId := d.Id()

	req := pipes.UpdateServiceAccountUserRequest{}
	var hasUpdate bool
	if d.HasChange("title") {
		req.SetTitle(d.Get("title").(string))
		hasUpdate = true
	}
	if d.HasChange("description") {
		req.SetDescription(d.Get("description").(string))
		hasUpdate = true
	}
	if hasUpdate {
		resp, r, err := client.APIClient.OrgServiceAccounts.Update(ctx, orgHandle, saId).Body(req).Execute()
		if err != nil {
			return diag.Errorf("resourceOrganizationServiceAccountUpdate. Update org service account error: %v", decodeResponse(r))
		}
		setOrganizationServiceAccountFields(d, &resp)
	}

	return diags
}

func resourceOrganizationServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)
	var diags diag.Diagnostics

	orgHandle := d.Get("organization_handle").(string)
	saId := d.Id()

	r, err := client.APIClient.OrgServiceAccounts.Delete(ctx, orgHandle, saId).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationServiceAccountDelete. Delete org service account error: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}

func setOrganizationServiceAccountFields(d *schema.ResourceData, resp *pipes.User) {
	// Common fields from User model
	// Note: organization is provided via organization_handle attribute
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
