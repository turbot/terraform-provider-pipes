package pipes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipes-sdk-go"
)

func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,
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
			"organization": {
				Type:     schema.TypeString,
				Required: true,
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
			"identity_id": {
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
			"config_sensitive": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				ValidateFunc: validation.StringIsJSON,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_error_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_error_process_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_successful_update_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_successful_update_process_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_update_attempt_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"last_update_attempt_process_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var plugin, connHandle, configString, orgHandle string
	var config map[string]interface{}
	var err error

	// Organization is manadatory now since we no longer have user level connections
	if val, ok := d.GetOk("organization"); ok {
		orgHandle = val.(string)
	}

	if value, ok := d.GetOk("handle"); ok {
		connHandle = value.(string)
	}
	if value, ok := d.GetOk("plugin"); ok {
		plugin = value.(string)
	}

	// Parse config and config_sensitive (if provided)
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(value.(string))
	}
	var configSensitive map[string]interface{}
	if value, ok := d.GetRawConfig().AsValueMap()["config_sensitive"]; ok && !value.IsNull() {
		_, configSensitive = formatConnectionJSONString(value.AsString())
	}
	// Merge shallow: config as base, config_sensitive overrides
	mergedConfig := config
	if configSensitive != nil {
		mergedConfig = mergeShallow(mergedConfig, configSensitive)
	}

	req := pipes.CreateConnectionRequest{
		Handle: connHandle,
		Plugin: plugin,
	}

	if mergedConfig != nil {
		req.SetConfig(mergedConfig)
	}

	client := meta.(*PipesClient)
	var resp pipes.Connection
	var r *http.Response

	resp, r, err = client.APIClient.OrgConnections.Create(ctx, orgHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceConnectionCreate. Create connection api error  %v", decodeResponse(r))
	}

	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("organization", orgHandle)
	d.Set("type", resp.Type)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Status != nil {
		d.Set("status", resp.Status)
	}
	if resp.LastErrorAt != nil {
		d.Set("last_error_at", resp.LastErrorAt)
	}
	if resp.LastErrorProcessId != nil {
		d.Set("last_error_process_id", resp.LastErrorProcessId)
	}
	if resp.LastSuccessfulUpdateAt != nil {
		d.Set("last_successful_update_at", resp.LastSuccessfulUpdateAt)
	}
	if resp.LastSuccessfulUpdateProcessId != nil {
		d.Set("last_successful_update_process_id", resp.LastSuccessfulUpdateProcessId)
	}
	if resp.LastUpdateAttemptAt != nil {
		d.Set("last_update_attempt_at", resp.LastUpdateAttemptAt)
	}
	if resp.LastUpdateAttemptProcessId != nil {
		d.Set("last_update_attempt_process_id", resp.LastUpdateAttemptProcessId)
	}
	// Set config from API response if present
	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err == nil && configString != "" && configString != "null" {
			d.Set("config", configString)
		}
	}
	// format "OrganizationHandle/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, *resp.Handle))

	return diags
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionHandle, orgHandle string
	var diags diag.Diagnostics
	var err error
	var r *http.Response
	var resp pipes.Connection
	id := d.Id()

	// format "OrganizationHandle/ConnectionHandle"
	ids := strings.Split(id, "/")
	orgHandle = ids[0]
	connectionHandle = ids[1]

	if orgHandle == "" {
		return diag.Errorf("resourceConnectionRead. Organization handle not present.")
	}
	if connectionHandle == "" {
		return diag.Errorf("resourceConnectionRead. Connection handle not present.")
	}

	resp, r, err = client.APIClient.OrgConnections.Get(context.Background(), orgHandle, connectionHandle).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection (%s) not found", connectionHandle),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceConnectionRead. Get connection error: %v", decodeResponse(r))
	}

	// Convert config to string
	var configString string
	configString, err = mapToJSONString(resp.GetConfig())
	if err != nil {
		return diag.Errorf("resourceOrganizationConnectionRead. Error converting config to string: %v", err)
	}

	// assign results back into ResourceData
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("organization", orgHandle)
	d.Set("type", resp.Type)
	d.Set("config", configString)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("handle", resp.Handle)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Status != nil {
		d.Set("status", resp.Status)
	}
	if resp.LastErrorAt != nil {
		d.Set("last_error_at", resp.LastErrorAt)
	}
	if resp.LastErrorProcessId != nil {
		d.Set("last_error_process_id", resp.LastErrorProcessId)
	}
	if resp.LastSuccessfulUpdateAt != nil {
		d.Set("last_successful_update_at", resp.LastSuccessfulUpdateAt)
	}
	if resp.LastSuccessfulUpdateProcessId != nil {
		d.Set("last_successful_update_process_id", resp.LastSuccessfulUpdateProcessId)
	}
	if resp.LastUpdateAttemptAt != nil {
		d.Set("last_update_attempt_at", resp.LastUpdateAttemptAt)
	}
	if resp.LastUpdateAttemptProcessId != nil {
		d.Set("last_update_attempt_process_id", resp.LastUpdateAttemptProcessId)
	}
	// format "OrganizationHandle/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, *resp.Handle))

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var orgHandle, configString string
	var r *http.Response
	var resp pipes.Connection
	var err error
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Organization is manadatory now since we no longer have user level connections
	if val, ok := d.GetOk("organization"); ok {
		orgHandle = val.(string)
	}

	oldConnectionHandle, newConnectionHandle := d.GetChange("handle")
	if newConnectionHandle.(string) == "" {
		return diag.Errorf("handle must be configured")
	}

	// Parse config and config_sensitive (if provided)
	if value, ok := d.GetOk("config"); ok {
		configString, config = formatConnectionJSONString(value.(string))
	}
	var configSensitive map[string]interface{}
	if value, ok := d.GetRawConfig().AsValueMap()["config_sensitive"]; ok && !value.IsNull() {
		_, configSensitive = formatConnectionJSONString(value.AsString())
	}
	// Merge shallow: config as base, config_sensitive overrides
	mergedConfig := config
	if configSensitive != nil {
		mergedConfig = mergeShallow(mergedConfig, configSensitive)
	}

	req := pipes.UpdateConnectionRequest{Handle: types.String(newConnectionHandle.(string))}
	if mergedConfig != nil {
		req.SetConfig(mergedConfig)
	}

	resp, r, err = client.APIClient.OrgConnections.Update(context.Background(), orgHandle, oldConnectionHandle.(string)).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceConnectionUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("handle", resp.Handle)
	d.Set("organization", orgHandle)
	d.Set("connection_id", resp.Id)
	d.Set("identity_id", resp.IdentityId)
	d.Set("type", resp.Type)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	d.Set("plugin", *resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	// Set config from API response if present
	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err == nil && configString != "" && configString != "null" {
			d.Set("config", configString)
		}
	}
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Status != nil {
		d.Set("status", resp.Status)
	}
	if resp.LastErrorAt != nil {
		d.Set("last_error_at", resp.LastErrorAt)
	}
	if resp.LastErrorProcessId != nil {
		d.Set("last_error_process_id", resp.LastErrorProcessId)
	}
	if resp.LastSuccessfulUpdateAt != nil {
		d.Set("last_successful_update_at", resp.LastSuccessfulUpdateAt)
	}
	if resp.LastSuccessfulUpdateProcessId != nil {
		d.Set("last_successful_update_process_id", resp.LastSuccessfulUpdateProcessId)
	}
	if resp.LastUpdateAttemptAt != nil {
		d.Set("last_update_attempt_at", resp.LastUpdateAttemptAt)
	}
	if resp.LastUpdateAttemptProcessId != nil {
		d.Set("last_update_attempt_process_id", resp.LastUpdateAttemptProcessId)
	}
	// format "OrganizationHandle/ConnectionHandle"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, *resp.Handle))

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionHandle, orgHandle string

	// Organization is manadatory now since we no longer have user level connections
	if val, ok := d.GetOk("organization"); ok {
		orgHandle = val.(string)
	}
	if value, ok := d.GetOk("handle"); ok {
		connectionHandle = value.(string)
	}

	var err error
	var r *http.Response

	_, r, err = client.APIClient.OrgConnections.Delete(ctx, orgHandle, connectionHandle).Execute()
	if err != nil {
		return diag.Errorf("resourceConnectionDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}

// config is a json string
// apply standard formatting to old and new data then compare
func connectionJSONStringsEqual(k, old, new string, d *schema.ResourceData) bool {
	if old == "" || new == "" {
		return false
	}
	oldFormatted, _ := formatConnectionJSONString(old)
	newFormatted, _ := formatConnectionJSONString(new)
	return oldFormatted == newFormatted
}

// apply standard formatting to a json string by unmarshalling into a map then marshalling back to JSON
func formatConnectionJSONString(body string) (string, map[string]interface{}) {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, []byte(body))
	if err != nil {
		return body, nil
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(buffer.Bytes(), &data); err != nil {
		// ignore error and just return original body
		return body, nil
	}

	body, err = mapToJSONString(data)
	if err != nil {
		// ignore error and just return original body
		return body, data
	}
	return body, data
}
