package pipes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipes-sdk-go"
)

func resourceWorkspaceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceConnectionCreate,
		ReadContext:   resourceWorkspaceConnectionRead,
		UpdateContext: resourceWorkspaceConnectionUpdate,
		DeleteContext: resourceWorkspaceConnectionDelete,
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
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
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
			"config_sensitive": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				WriteOnly:    true,
				ValidateFunc: validation.StringIsJSON,
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
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceWorkspaceConnectionV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceWorkspaceConnectionStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceWorkspaceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var workspaceHandle, plugin, connHandle, configString, parentId string
	var config map[string]interface{}
	var err error

	// Get details about the workspace where the connection would be created
	if val, ok := d.GetOk("workspace"); ok {
		workspaceHandle = val.(string)
	}

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

	// Parse config and config_sensitive (if provided)
	if value, ok := d.GetOk("config"); ok {
		_, config = formatConnectionJSONString(value.(string))
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

	// Pass the parent_id if its set
	if parentId != "" {
		req.SetParentId(parentId)
	}

	// Pass the config if its set
	if mergedConfig != nil {
		req.SetConfig(mergedConfig)
	}

	client := meta.(*PipesClient)
	var resp pipes.Connection
	var r *http.Response

	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionCreate. getUserHandler error %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnections.Create(ctx, actorHandle, workspaceHandle).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnections.Create(ctx, orgHandle, workspaceHandle).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionCreate. Create connection api error  %v", decodeResponse(r))
	}

	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionCreate. Error converting config to string: %v", err)
		}
	}

	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	if configString != "" && configString != "null" {
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
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	// ID Format
	// User workspace connection - WorkspaceHandle/ConnectionHandle
	// Org workspace connection - OrgHandle/WorkspaceHandle/ConnectionHandle
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Handle))

	}

	return diags
}

func resourceWorkspaceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionHandle, orgHandle, workspaceHandle, configString string
	var diags diag.Diagnostics
	var isUser = false

	// ID formats
	// User workspace connection - "WorkspaceHandle/ConnectionHandle"
	// Org workspace connection - "OrganizationHandle/WorkspaceHandle/ConnectionHandle"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <org-handle>/<workspace-handle>/<connection-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		connectionHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		connectionHandle = idParts[1]
	}

	if workspaceHandle == "" {
		return diag.Errorf("resourceWorkspaceConnectionRead. Workspace information not present.")
	}
	if connectionHandle == "" {
		return diag.Errorf("resourceWorkspaceConnectionRead. Connection handle not present.")
	}

	var resp pipes.WorkspaceConnection
	var r *http.Response
	var err error
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionCreate. getUserHandler error %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnections.Get(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnections.Get(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionCreate. Create connection api error  %v", decodeResponse(r))
	}

	// Convert config to string
	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceOrganizationConnectionRead. Error converting config to string: %v", err)
		}
	}

	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	if configString != "" && configString != "null" {
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
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	// ID Format
	// User workspace connection - WorkspaceHandle/ConnectionHandle
	// Org workspace connection - OrgHandle/WorkspaceHandle/ConnectionHandle
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Handle))

	}

	return diags
}

func resourceWorkspaceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var workspaceHandle, configString string
	var config map[string]interface{}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Get details about the workspace where the integration would be created
	if val, ok := d.GetOk("workspace"); ok {
		workspaceHandle = val.(string)
	}

	oldConnectionHandle, newConnectionHandle := d.GetChange("handle")
	if newConnectionHandle.(string) == "" {
		return diag.Errorf("handle must be configured")
	}

	// Parse config and config_sensitive (if provided)
	if value, ok := d.GetOk("config"); ok {
		_, config = formatConnectionJSONString(value.(string))
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
	if ok := d.HasChange("parent_id"); ok {
		if value, ok := d.GetOk("parent_id"); ok {
			req.SetParentId(value.(string))
		}
	}
	if ok := d.HasChange("config_source"); ok {
		if value, ok := d.GetOk("config_source"); ok {
			req.SetConfigSource(pipes.ConnectionConfigSource(value.(string)))
		}
	}
	if ok := d.HasChange("credential_source"); ok {
		if value, ok := d.GetOk("credential_source"); ok {
			req.SetCredentialSource(pipes.ConnectionCredentialSource(value.(string)))
		}
	}

	var resp pipes.Connection
	var err error
	var r *http.Response
	isUser, orgHandle := isUserConnection(d)
	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionUpdate. getUserHandler error %v", decodeResponse(r))
		}
		resp, r, err = client.APIClient.UserWorkspaceConnections.Update(ctx, actorHandle, workspaceHandle, oldConnectionHandle.(string)).Request(req).Execute()
	} else {
		resp, r, err = client.APIClient.OrgWorkspaceConnections.Update(ctx, orgHandle, workspaceHandle, oldConnectionHandle.(string)).Request(req).Execute()
	}
	if err != nil {
		return diag.Errorf("resourceWorkspaceConnectionUpdate. Create connection api error  %v", decodeResponse(r))
	}

	if resp.GetConfig() != nil {
		configString, err = mapToJSONString(resp.GetConfig())
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionUpdate. Error converting config to string: %v", err)
		}
	}

	d.Set("connection_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("handle", resp.Handle)
	d.Set("plugin", resp.Plugin)
	d.Set("plugin_version", resp.PluginVersion)
	d.Set("type", resp.Type)
	if configString != "" && configString != "null" {
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
	d.Set("organization", orgHandle)
	d.Set("workspace", workspaceHandle)
	// ID Format
	// User workspace connection - WorkspaceHandle/ConnectionHandle
	// Org workspace connection - OrgHandle/WorkspaceHandle/ConnectionHandle
	if isUser {
		d.SetId(fmt.Sprintf("%s/%s", workspaceHandle, *resp.Handle))
	} else {
		d.SetId(fmt.Sprintf("%s/%s/%s", orgHandle, workspaceHandle, *resp.Handle))

	}

	return diags
}

func resourceWorkspaceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, workspaceHandle, connectionHandle string
	var isUser = false

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 && len(idParts) > 3 {
		return diag.Errorf("unexpected format of ID (%q), expected <org-handle>/<workspace-handle>/<connection-handle>", d.Id())
	}

	if len(idParts) == 3 {
		orgHandle = idParts[0]
		workspaceHandle = idParts[1]
		connectionHandle = idParts[2]
	} else if len(idParts) == 2 {
		isUser = true
		workspaceHandle = idParts[0]
		connectionHandle = idParts[1]
	}

	log.Printf("\n[DEBUG] Deleting Workspace Connection: %s", fmt.Sprintf("%s/%s", workspaceHandle, connectionHandle))

	var err error
	var r *http.Response

	if isUser {
		var actorHandle string
		actorHandle, r, err = getUserHandler(ctx, client)
		if err != nil {
			return diag.Errorf("resourceWorkspaceConnectionDelete. getUserHandler error: %v", decodeResponse(r))
		}
		_, r, err = client.APIClient.UserWorkspaceConnections.Delete(ctx, actorHandle, workspaceHandle, connectionHandle).Execute()
	} else {
		_, r, err = client.APIClient.OrgWorkspaceConnections.Delete(ctx, orgHandle, workspaceHandle, connectionHandle).Execute()
	}

	if err != nil {
		return diag.Errorf("error deleting workspace connection: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}

func resourceWorkspaceConnectionV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"connection_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_]{0,37}[a-z0-9]?$`), "Handle must be between 1 and 39 characters, and may only contain alphanumeric characters or single underscores, cannot start with a number or underscore and cannot end with an underscore."),
			},
			"workspace_handle": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9]{1,23}$`), "Handle must be between 1 and 23 characters, and may only contain alphanumeric characters."),
			},
			"organization": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
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
			"identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"workspace_created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_hive": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_identity_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_public_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"workspace_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"workspace_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWorkspaceConnectionStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	// Throw error mentioning that the user needs to be migrate the connection to a different resource type
	return rawState, fmt.Errorf("`pipes_workspace_connection` has been moved to resource type `pipes_workspace_schema`. Please move existing resources to the new resource type. For more information, refer to the documentation")
}
