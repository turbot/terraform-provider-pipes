package pipes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceTenantConnectionFolderPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantConnectionFolderPermissionCreate,
		ReadContext:   resourceTenantConnectionFolderPermissionRead,
		UpdateContext: resourceTenantConnectionFolderPermissionUpdate,
		DeleteContext: resourceTenantConnectionFolderPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"permission_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
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
			"connection_folder_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				ConflictsWith: []string{"identity_handle", "workspace_handle"},
			},
			"identity_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				ConflictsWith: []string{"tenant_handle"},
			},
			"workspace_handle": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      false,
				RequiredWith:  []string{"identity_handle"},
				ConflictsWith: []string{"tenant_handle"},
			},
		},
	}
}

func resourceTenantConnectionFolderPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var tenantHandle, identityHandle, workspaceHandle, connectionFolderId string
	var err error

	// When attaching a workspace schema, we can pass in a connection folder id, connection handle or aggregator handle
	// Its already verified as part of schema validation rules that only one of these can be defined in configuration
	if val, ok := d.GetOk("tenant_handle"); ok {
		tenantHandle = val.(string)
	}
	if val, ok := d.GetOk("identity_handle"); ok {
		identityHandle = val.(string)
	}
	if val, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = val.(string)
	}
	if val, ok := d.GetOk("connection_folder_id"); ok {
		connectionFolderId = val.(string)
	}

	// Frame the request object
	req := pipes.CreatePermissionRequest{
		TenantHandle:    &tenantHandle,
		IdentityHandle:  &identityHandle,
		WorkspaceHandle: &workspaceHandle,
	}

	client := meta.(*PipesClient)
	var resp pipes.Permission
	var r *http.Response

	// Create permission for connection
	resp, r, err = client.APIClient.TenantConnectionFolders.CreatePermission(ctx, connectionFolderId).Request(req).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error creating tenant connection folder permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.Handle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.Handle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.Handle)
	}
	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	d.Set("connection_folder_id", connectionFolderId)
	d.SetId(fmt.Sprintf("%s/%s", connectionFolderId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionFolderId, permissionId string

	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <connection-handle>/<permission-id>", d.Id())
	}

	connectionFolderId = idParts[0]
	permissionId = idParts[1]

	var resp pipes.Permission
	var err error
	var r *http.Response

	resp, r, err = client.APIClient.TenantConnectionFolders.GetPermission(ctx, connectionFolderId, permissionId).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error getting tenant connection folder permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.Handle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.Handle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.Handle)
	}
	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	d.Set("connection_folder_id", connectionFolderId)
	d.SetId(fmt.Sprintf("%s/%s", connectionFolderId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderPermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var tenantHandle, identityHandle, workspaceHandle, connectionFolderId, permissionId string
	var err error

	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <connection-folder-id>/<permission-id>", d.Id())
	}

	permissionId = idParts[1]

	// When attaching a workspace schema, we can pass in a connection folder id, connection handle or aggregator handle
	// Its already verified as part of schema validation rules that only one of these can be defined in configuration
	if val, ok := d.GetOk("tenant_handle"); ok {
		tenantHandle = val.(string)
	}
	if val, ok := d.GetOk("identity_handle"); ok {
		identityHandle = val.(string)
	}
	if val, ok := d.GetOk("workspace_handle"); ok {
		workspaceHandle = val.(string)
	}
	if val, ok := d.GetOk("connection_folder_id"); ok {
		connectionFolderId = val.(string)
	}

	// Frame the request object
	req := pipes.UpdatePermissionRequest{
		TenantHandle:    &tenantHandle,
		IdentityHandle:  &identityHandle,
		WorkspaceHandle: &workspaceHandle,
	}

	client := meta.(*PipesClient)
	var resp pipes.Permission
	var r *http.Response

	// Create permission for connection
	resp, r, err = client.APIClient.TenantConnectionFolders.UpdatePermission(ctx, connectionFolderId, permissionId).Request(req).Execute()
	// Error check
	if err != nil {
		return diag.Errorf("error updating tenant connection permission: %v", decodeResponse(r))
	}

	// Set property values
	d.Set("permission_id", resp.Id)
	d.Set("connection_id", resp.ConnectionId)
	d.Set("tenant_id", resp.TenantId)
	d.Set("identity_id", resp.IdentityId)
	d.Set("workspace_id", resp.WorkspaceId)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	if resp.Tenant != nil {
		d.Set("tenant_handle", resp.Tenant.Handle)
	}
	if resp.Identity != nil {
		d.Set("identity_handle", resp.Identity.Handle)
	}
	if resp.Workspace != nil {
		d.Set("workspace_handle", resp.Workspace.Handle)
	}
	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	d.Set("connection_folder_id", connectionFolderId)
	d.SetId(fmt.Sprintf("%s/%s", connectionFolderId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionFolderId, permissionId string

	// ID formats
	// Tenant Connection - "ConnectionFolderId/PermissionId"
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return diag.Errorf("unexpected format of ID (%q), expected <connection-handle>/<permission-id>", d.Id())
	}

	connectionFolderId = idParts[0]
	permissionId = idParts[1]

	var err error
	var r *http.Response

	_, r, err = client.APIClient.TenantConnectionFolders.DeletePermission(ctx, connectionFolderId, permissionId).Execute()
	if err != nil {
		return diag.Errorf("error deleting permission from tenant connection: %v", decodeResponse(r))
	}
	d.SetId("")

	return diags
}
