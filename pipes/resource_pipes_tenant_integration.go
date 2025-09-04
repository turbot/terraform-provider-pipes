package pipes

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/turbot/pipes-sdk-go"
)

func resourceTenantIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantIntegrationCreate,
		ReadContext:   resourceTenantIntegrationRead,
		UpdateContext: resourceTenantIntegrationUpdate,
		DeleteContext: resourceTenantIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
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
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: IntegrationJSONStringsEqual,
				ConflictsWith:    []string{"config_wo"},
			},
			"config_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				ValidateFunc:  validation.StringIsJSON,
				ConflictsWith: []string{"config"},
				RequiredWith:  []string{"config_wo_version"},
			},
			"config_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"config_wo"},
			},
			"github_installation_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"pipeline_id": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceTenantIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var integrationType, integrationHandle string
	var configString string
	var config map[string]interface{}
	var err error

	if value, ok := d.GetOk("handle"); ok {
		integrationHandle = value.(string)
	}
	if value, ok := d.GetOk("type"); ok {
		integrationType = value.(string)
	}

	// Parse config OR config_wo (if provided)
	var writeConfig bool
	if value, ok := d.GetOk("config"); ok {
		writeConfig = true
		configString, config = FormatIntegrationJSONString(value.(string))
	}
	if value, ok := d.GetRawConfig().AsValueMap()["config_wo"]; ok && !value.IsNull() {
		_, config = FormatIntegrationJSONString(value.AsString())
	}

	req := pipes.CreateIntegrationRequest{
		Handle: integrationHandle,
		Type:   pipes.IntegrationType(integrationType),
	}

	if config != nil {
		req.SetConfig(config)
	}

	client := meta.(*PipesClient)
	var resp pipes.Integration
	var r *http.Response

	resp, r, err = client.APIClient.TenantIntegrations.Create(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantIntegrationCreate. Create integration api error  %v", decodeResponse(r))
	}

	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceTenantIntegrationCreate. Error converting config to string: %v", err)
		}
	}
	if writeConfig && configString != "" && configString != "null" {
		d.Set("config", configString)
	}
	d.Set("github_installation_id", resp.GithubInstallationId)
	d.Set("pipeline_id", resp.PipelineId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The integration is being created at a custom tenant level
	// hence the id would be of format "TenantHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Handle))

	return diags
}

func resourceTenantIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var integrationId, tenantId, configString string
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.Integration
	id := d.Id()

	// For backward-compatibility, we see whether the id contains : or /
	separator := "/"
	if strings.Contains(id, ":") {
		separator = ":"
	}

	// The id consists of parts in thr format "TenantHandle/IntegrationHandle"
	ids := strings.Split(id, separator)
	if len(ids) == 2 {
		tenantId = ids[0]
		integrationId = ids[1]
	}

	if tenantId == "" {
		return diag.Errorf("resourceTenantIntegrationRead. Tenant information not present.")
	}
	if integrationId == "" {
		return diag.Errorf("resourceTenantIntegrationRead. Integration information not present.")
	}

	resp, r, err = client.APIClient.TenantIntegrations.Get(ctx, integrationId).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Integration (%s) not found", integrationId),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceTenantIntegrationRead. Get integration error: %v", decodeResponse(r))
	}

	// Convert config to string
	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceTenantIntegrationRead. Error converting config to string: %v", err)
		}
	}
	_, writeConfig := d.GetOk("config")

	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	if writeConfig && configString != "" && configString != "null" {
		d.Set("config", configString)
	}
	d.Set("github_installation_id", resp.GithubInstallationId)
	d.Set("pipeline_id", resp.PipelineId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The integration is being created at a custom tenant level
	// hence the id would be of format "TenantHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Handle))

	return diags
}

func resourceTenantIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var state string
	var configString string
	var r *http.Response
	var resp pipes.Integration
	var err error
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	o, n := d.GetChange("handle")
	oldHandle := o.(string)
	newHandle := n.(string)
	if newHandle == "" {
		return diag.Errorf("handle must be configured")
	}
	if value, ok := d.GetOk("state"); ok {
		state = value.(string)
	}

	// Parse config OR config_wo (if provided)
	var writeConfig bool
	if value, ok := d.GetOk("config"); ok {
		writeConfig = true
		configString, config = FormatIntegrationJSONString(value.(string))
	}
	if value, ok := d.GetRawConfig().AsValueMap()["config_wo"]; ok && !value.IsNull() {
		_, config = FormatIntegrationJSONString(value.AsString())
	}

	req := pipes.UpdateIntegrationRequest{
		Handle: &newHandle,
		State:  (*pipes.IntegrationState)(&state),
	}

	if config != nil {
		req.SetConfig(config)
	}

	resp, r, err = client.APIClient.TenantIntegrations.Update(ctx, oldHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantIntegrationUpdate. Update integration error: %v", decodeResponse(r))
	}

	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceTenantIntegrationUpdate. Error converting config to string: %v", err)
		}
	}

	d.Set("integration_id", resp.Id)
		d.Set("tenant_id", resp.TenantId)
		d.Set("handle", resp.Handle)
		d.Set("type", resp.Type)
		d.Set("state", resp.State)
		d.Set("state_reason", resp.StateReason)
	if writeConfig && configString != "" && configString != "null" {
		d.Set("config", configString)
	}
		d.Set("github_installation_id", resp.GithubInstallationId)
	d.Set("pipeline_id", resp.PipelineId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The integration is being created at a custom tenant level
	// hence the id would be of format "TenantHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Handle))

	return diags
}

func resourceTenantIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var integrationHandle string
	if value, ok := d.GetOk("handle"); ok {
		integrationHandle = value.(string)
	}

	_, r, err := client.APIClient.TenantIntegrations.Delete(ctx, integrationHandle).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantIntegrationDelete. Delete integration error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
