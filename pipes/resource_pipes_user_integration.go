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

func resourceUserIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserIntegrationCreate,
		ReadContext:   resourceUserIntegrationRead,
		UpdateContext: resourceUserIntegrationUpdate,
		DeleteContext: resourceUserIntegrationDelete,
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
			"identity_id": {
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
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: IntegrationJSONStringsEqual,
			},
			"github_installation_id": {
				Type:     schema.TypeString,
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

func resourceUserIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var integrationType, integrationHandle string
	var configString string
	var config map[string]interface{}

	if value, ok := d.GetOk("handle"); ok {
		integrationHandle = value.(string)
	}
	if value, ok := d.GetOk("type"); ok {
		integrationType = value.(string)
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = FormatIntegrationJSONString(value.(string))
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

	// Get actor information from the Actors endpoint
	actorHandle, r, err := getUserHandler(ctx, client)
	if err != nil {
		return diag.Errorf("resourceUserIntegrationCreate. getUserHandler error  %v", decodeResponse(r))
	}

	// Create the integration for the user identity
	resp, r, err = client.APIClient.UserIntegrations.Create(ctx, actorHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceUserIntegrationCreate. Create integration api error  %v", decodeResponse(r))
	}

	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	if config != nil {
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
	// The integration is being created at the user level
	// hence the id would be of format "UserHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", actorHandle, resp.Handle))

	return diags
}

func resourceUserIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var integrationHandle, userHandle string
	var diags diag.Diagnostics
	id := d.Id()

	// Id is fof format "UserHandle/IntegrationHandle"
	ids := strings.Split(id, "/")
	if len(ids) == 2 {
		userHandle = ids[0]
		integrationHandle = ids[1]
	}

	if integrationHandle == "" {
		return diag.Errorf("resourceUserIntegrationRead. Integration details is not present.")
	}

	resp, r, err := client.APIClient.UserIntegrations.Get(context.Background(), userHandle, integrationHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Integration (%s) not found", integrationHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceUserIntegrationRead. Get integration error: %v", decodeResponse(r))
	}

	// Convert config to string
	configString, err := mapToJSONString(resp.GetConfig())
	if err != nil {
		return diag.Errorf("resourceTenantIntegrationRead. Error converting config to string: %v", err)
	}

	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	d.Set("config", configString)
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
	// The integration is being created at the user level
	// hence the id would be of format "UserHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", userHandle, resp.Handle))

	return diags
}

func resourceUserIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var state string
	var configString string
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	old, new := d.GetChange("handle")
	if new.(string) == "" {
		return diag.Errorf("handle must be configured")
	}
	if value, ok := d.GetOk("state"); ok {
		state = value.(string)
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = FormatIntegrationJSONString(value.(string))
	}

	oldHandle := old.(string)
	newHandle := new.(string)

	req := pipes.UpdateIntegrationRequest{
		Handle: &newHandle,
		State:  (*pipes.IntegrationState)(&state),
	}

	if config != nil {
		req.SetConfig(config)
	}

	// Get actor information from the Actors endpoint
	actorHandle, r, err := getUserHandler(ctx, client)
	if err != nil {
		return diag.Errorf("resourceUserIntegrationCreate. getUserHandler error  %v", decodeResponse(r))
	}

	resp, r, err := client.APIClient.UserIntegrations.Update(context.Background(), actorHandle, oldHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceUserIntegrationUpdate. Update integration error: %v", decodeResponse(r))
	}

	d.Set("integration_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("handle", resp.Handle)
	d.Set("type", resp.Type)
	d.Set("state", resp.State)
	d.Set("state_reason", resp.StateReason)
	if config != nil {
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
	// The integration is being created at the user level
	// hence the id would be of format "UserHandle/IntegrationHandle"
	d.SetId(fmt.Sprintf("%s/%s", actorHandle, resp.Handle))

	return diags
}

func resourceUserIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var integrationHandle string
	if value, ok := d.GetOk("handle"); ok {
		integrationHandle = value.(string)
	}

	var actorHandle string
	actorHandle, r, err := getUserHandler(ctx, client)
	if err != nil {
		return diag.Errorf("resourceUserIntegrationDelete. getUserHandler error: %v", decodeResponse(r))
	}

	_, r, err = client.APIClient.UserIntegrations.Delete(ctx, actorHandle, integrationHandle).Execute()
	if err != nil {
		return diag.Errorf("resourceUserIntegrationDelete. Delete integration error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
