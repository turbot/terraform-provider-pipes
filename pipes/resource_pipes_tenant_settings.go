package pipes

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/pipes-sdk-go"
)

func resourceTenantSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantSettingsUpdate,
		ReadContext:   resourceTenantSettingsRead,
		UpdateContext: resourceTenantSettingsUpdate,
		DeleteContext: resourceTenantSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			// Timeouts are in seconds as per UpdateTenantSettingsRequest
			"cli_session_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"console_session_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"max_token_expiration": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},

			// Login method states
			"login_email_state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			"login_github_state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			"login_google_state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			// SAML settings
			"login_saml_state": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			"login_saml_certificate": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"login_saml_issuer": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_saml_sso_url": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Personal workspaces and postgres endpoint states
			"personal_workspaces": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},
			"postgres_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"enabled", "disabled"}, false),
			},

			// Arrays
			"user_provisioning": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_provisioning_permitted_domains": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspace_snapshot_permitted_visibility": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Metadata
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
		},
	}
}

func resourceTenantSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	resp, r, err := client.APIClient.Tenants.GetSettings(ctx).Execute()
	if err != nil {
		log.Printf("\n[WARN] Tenant settings not found or not accessible")
		d.SetId("")
		return diag.Errorf("error reading tenant settings: %v", decodeResponse(r))
	}

	resourceTenantSettingsPopulateFromResponse(d, resp)

	return diags
}

func resourceTenantSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*PipesClient)

	req := pipes.UpdateTenantSettingsRequest{}
	current, r, err := client.APIClient.Tenants.GetSettings(ctx).Execute()
	if err != nil {
		log.Printf("\n[WARN] Tenant settings not found or not accessible: %v", decodeResponse(r))
		return diags
	}

	if v, ok := d.GetOk("cli_session_timeout"); ok {
		val := int32(v.(int))
		if current.CliSessionTimeout == nil || *current.CliSessionTimeout != val {
			req.CliSessionTimeout = &val
		}
	}
	if v, ok := d.GetOk("console_session_timeout"); ok {
		val := int32(v.(int))
		if current.ConsoleSessionTimeout == nil || *current.ConsoleSessionTimeout != val {
			req.ConsoleSessionTimeout = &val
		}
	}

	// max_token_expiration is a special case as we want to handle 0
	if v := d.Get("max_token_expiration"); v != nil {
		val := int32(v.(int))
		if current.MaxTokenExpiration != val {
			req.MaxTokenExpiration = &val
		}
	}

	// Login states
	if v, ok := d.GetOk("login_email_state"); ok {
		if current.LoginEmail.State != v.(string) {
			req.LoginEmail = &pipes.UpdateTenantLoginSettings{State: v.(string)}
		}
	}
	if v, ok := d.GetOk("login_github_state"); ok {
		if current.LoginGithub.State != v.(string) {
			req.LoginGithub = &pipes.UpdateTenantLoginSettings{State: v.(string)}
		}
	}
	if v, ok := d.GetOk("login_google_state"); ok {
		if current.LoginGoogle.State != v.(string) {
			req.LoginGoogle = &pipes.UpdateTenantLoginSettings{State: v.(string)}
		}
	}
	if v, ok := d.GetOk("login_saml_state"); ok {
		state := v.(string)
		saml := pipes.UpdateTenantSamlLoginSettings{State: state}
		if v, ok := d.GetOk("login_saml_certificate"); ok {
			cert := v.(string)
			saml.Certificate = &cert
		}
		if v, ok := d.GetOk("login_saml_issuer"); ok {
			issuer := v.(string)
			saml.Issuer = &issuer
		}
		if v, ok := d.GetOk("login_saml_sso_url"); ok {
			sso := v.(string)
			saml.SsoUrl = &sso
		}
		req.LoginSaml = &saml
	}

	if v, ok := d.GetOk("personal_workspaces"); ok {
		pw := pipes.TenantPersonalWorkspaces(v.(string))
		if current.PersonalWorkspaces != pw {
			req.PersonalWorkspaces = &pw
		}
	}
	if v, ok := d.GetOk("postgres_endpoint"); ok {
		pe := pipes.PostgresEndpointState(v.(string))
		if current.PostgresEndpoint == nil || *current.PostgresEndpoint != pe {
			req.PostgresEndpoint = &pe
		}
	}

	if v, ok := d.GetOk("user_provisioning"); ok {
		list, err := convertToStringArray(v.([]interface{}))
		if err != nil {
			return diag.Errorf("error converting user_provisioning to string array: %v", err)
		}
		req.UserProvisioning = &list
	}
	if v, ok := d.GetOk("user_provisioning_permitted_domains"); ok {
		list, err := convertToStringArray(v.([]interface{}))
		if err != nil {
			return diag.Errorf("error converting user_provisioning_permitted_domains to string array: %v", err)
		}
		req.UserProvisioningPermittedDomains = &list
	}
	if v, ok := d.GetOk("workspace_snapshot_permitted_visibility"); ok {
		list, err := convertToStringArray(v.([]interface{}))
		if err != nil {
			return diag.Errorf("error converting workspace_snapshot_permitted_visibility to string array: %v", err)
		}
		req.WorkspaceSnapshotPermittedVisibility = &list
	}

	resp, r, err := client.APIClient.Tenants.UpdateSettings(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("error updating tenant settings: %v", decodeResponse(r))
	}

	resourceTenantSettingsPopulateFromResponse(d, resp)

	return diags
}

func resourceTenantSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// No API to delete settings; they are intrinsic to the tenant.
	// We leave current settings as-is and simply remove from state.
	var diags diag.Diagnostics
	d.SetId("")
	return diags
}

func resourceTenantSettingsPopulateFromResponse(d *schema.ResourceData, resp pipes.TenantSettings) {
	// tokens
	if resp.CliSessionTimeout != nil {
		d.Set("cli_session_timeout", *resp.CliSessionTimeout)
	}
	if resp.ConsoleSessionTimeout != nil {
		d.Set("console_session_timeout", *resp.ConsoleSessionTimeout)
	}
	d.Set("max_token_expiration", resp.MaxTokenExpiration)

	// auth states
	d.Set("login_email_state", resp.LoginEmail.State)
	d.Set("login_github_state", resp.LoginGithub.State)
	d.Set("login_google_state", resp.LoginGoogle.State)
	d.Set("login_saml_state", resp.LoginSaml.State)
	if resp.LoginSaml.Certificate != nil {
		d.Set("login_saml_certificate", *resp.LoginSaml.Certificate)
	}
	if resp.LoginSaml.Issuer != nil {
		d.Set("login_saml_issuer", *resp.LoginSaml.Issuer)
	}
	if resp.LoginSaml.SsoUrl != nil {
		d.Set("login_saml_sso_url", *resp.LoginSaml.SsoUrl)
	}

	// workspace configuration enums
	d.Set("personal_workspaces", string(resp.PersonalWorkspaces))
	if resp.PostgresEndpoint != nil {
		d.Set("postgres_endpoint", string(*resp.PostgresEndpoint))
	}

	// arrays
	d.Set("user_provisioning", resp.UserProvisioning)
	d.Set("user_provisioning_permitted_domains", resp.UserProvisioningPermittedDomains)
	d.Set("workspace_snapshot_permitted_visibility", resp.WorkspaceSnapshotPermittedVisibility)

	// metadata
	d.Set("created_at", resp.CreatedAt)
	if resp.UpdatedAt != nil {
		d.Set("updated_at", *resp.UpdatedAt)
	}
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)

	d.SetId("tenant/settings")
}
