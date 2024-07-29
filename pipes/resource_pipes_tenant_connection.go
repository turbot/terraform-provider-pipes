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
	"github.com/turbot/go-kit/types"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceTenantConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantConnectionCreate,
		ReadContext:   resourceTenantConnectionRead,
		UpdateContext: resourceTenantConnectionUpdate,
		DeleteContext: resourceTenantConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"connection_id": {
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
			"plugin": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plugin_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: connectionJSONStringsEqual,
			},
			"config_source": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"credential_source": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"handle_dynamic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_by_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trunk": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
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

func resourceTenantConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var plugin, connHandle, configString, parentId string
	var config map[string]interface{}
	var err error

	// Get general information about the connection to be created
	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}
	if value, ok := d.GetOk("parent_id"); ok {
		parentId = value.(string)
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(value.(string))
	}

	req := pipes.CreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	// Pass the parent_id if its set
	if parentId != "" {
		req.SetParentId(parentId)
	}

	// Pass the config if its set
	if config != nil {
		req.SetConfig(config)
	}

	client := meta.(*PipesClient)
	var resp pipes.Connection
	var r *http.Response

	resp, r, err = client.APIClient.TenantConnections.Create(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionCreate. Create connection api error  %v", decodeResponse(r))
	}

	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	if config != nil {
		d.Set("config", configString)
	}
	d.Set("config_source", resp.ConfigSource)
	d.Set("credential_source", resp.CredentialSource)
	d.Set("handle_mode", resp.HandleMode)
	d.Set("handle_dynamic", resp.HandleDynamic)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The connection is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, *resp.Handle))

	return diags
}

func resourceTenantConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionHandle, tenantId string
	var diags diag.Diagnostics

	// Its a tenant level connection so the id would be of format "TenantId/ConnectionHandle"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		tenantId = ids[0]
		connectionHandle = ids[1]
	}

	if tenantId == "" {
		return diag.Errorf("resourceTenantConnectionRead. Tenant information not present.")
	}
	if connectionHandle == "" {
		return diag.Errorf("resourceTenantConnectionRead. Connection handle not present.")
	}

	resp, r, err := client.APIClient.TenantConnections.Get(context.Background(), connectionHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection (%s) not found", connectionHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceTenantConnectionRead. Get connection error: %v", decodeResponse(r))
	}

	// Convert config to string
	var configString string
	configString, err = mapToJSONString(resp.GetConfig())
	if err != nil {
		return diag.Errorf("resourceTenantConnectionRead. Error converting config to string: %v", err)
	}

	// assign results back into ResourceData
	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	d.Set("config", configString)
	d.Set("config_source", resp.ConfigSource)
	d.Set("credential_source", resp.CredentialSource)
	d.Set("handle_mode", resp.HandleMode)
	d.Set("handle_dynamic", resp.HandleDynamic)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The connection is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, *resp.Handle))

	return diags
}

func resourceTenantConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var configString string
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	oldConnectionHandle, newConnectionHandle := d.GetChange("handle")
	if newConnectionHandle.(string) == "" {
		return diag.Errorf("handle must be configured")
	}

	// save the formatted data: this is to ensure the acceptance tests behave in a consistent way regardless of the ordering of the json data
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(value.(string))
	}

	req := pipes.UpdateConnectionRequest{Handle: types.String(newConnectionHandle.(string))}
	if config != nil {
		req.SetConfig(config)
	}
	if ok := d.HasChange("parent_id"); ok {
		if value, ok := d.GetOk("parent_id"); ok {
			req.SetParentId(value.(string))
		}
	}
	if ok := d.HasChange("config_source"); ok {
		if value, ok := d.GetOk("config_source"); ok {
			req.SetConfigSource(value.(string))
		}
	}
	if ok := d.HasChange("credential_source"); ok {
		if value, ok := d.GetOk("credential_source"); ok {
			req.SetCredentialSource(value.(string))
		}
	}

	resp, r, err := client.APIClient.TenantConnections.Update(context.Background(), oldConnectionHandle.(string)).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	if config != nil {
		d.Set("config", configString)
	}
	d.Set("config_source", resp.ConfigSource)
	d.Set("credential_source", resp.CredentialSource)
	d.Set("handle_mode", resp.HandleMode)
	d.Set("handle_dynamic", resp.HandleDynamic)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	// The connection is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, *resp.Handle))

	return diags
}

func resourceTenantConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionHandle string
	if value, ok := d.GetOk("handle"); ok {
		connectionHandle = value.(string)
	}

	_, r, err := client.APIClient.TenantConnections.Delete(ctx, connectionHandle).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
